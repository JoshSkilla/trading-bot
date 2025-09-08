package finnhub

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"

	cfg "github.com/joshskilla/trading-bot/internal/config"
	md "github.com/joshskilla/trading-bot/internal/marketdata"
	t "github.com/joshskilla/trading-bot/internal/types"
)

// Compile-time check to see if Client implements SampleProvider
var _ md.SampleProvider = (*Client)(nil)

// finnhub.Client is a Finnhub adapter
type Client struct {
	HTTP         *http.Client
	Token        string
	MaxStaleness time.Duration // max age of a valid cached sample

	// internal stream
	wsMu   sync.Mutex
	wsConn *websocket.Conn
	start  sync.Once

	// subscriptions (symbols)
	subs map[string]struct{}

	// cache
	cacheMu sync.RWMutex
	latest  map[string]t.Sample  // by symbol
	stamp   map[string]time.Time // last update time
}

func NewClient(token string) *Client {
	return &Client{
		HTTP:         &http.Client{Timeout: 10 * time.Second},
		Token:        token,
		MaxStaleness: 2 * time.Second,
		subs:         make(map[string]struct{}),
		latest:       make(map[string]t.Sample),
		stamp:        make(map[string]time.Time),
	}
}

// If streaming is enabled and cache is valid/fresh, serves from cache
// Else falls back to REST /quote
func (c *Client) FetchSample(ctx context.Context, asset t.Asset) (t.Sample, error) {
	// Ensure the asset is added to the stream if not already there
	c.wsMu.Lock()
	_, exists := c.subs[asset.Symbol]
	c.wsMu.Unlock()
	if !exists {
		_ = c.AddToStream(ctx, []t.Asset{asset})
	}
	// Serve from cache if fresh
	if sm, ok := c.getCached(asset.Symbol); ok && time.Since(sm.when) <= c.MaxStaleness {
		return sm.sample, nil
	}
	// fallback to REST
	return c.fetchSampleREST(ctx, asset)
}

// AddToStream starts the WS (on the first call) and subscribes any new assets.
// Safe to call multiple times; re-subs are ignored.
func (c *Client) AddToStream(ctx context.Context, assets []t.Asset) error {
	if len(assets) == 0 {
		return nil
	}

	// ensure connection (lazy start)
	if err := c.ensureConn(ctx); err != nil {
		return err
	}

	// subscribe any new symbols
	c.wsMu.Lock()
	defer c.wsMu.Unlock()

	for _, a := range assets {
		sym := a.Symbol
		if _, exists := c.subs[sym]; exists {
			continue
		}
		msg := fmt.Sprintf(`{"type":"subscribe","symbol":"%s"}`, sym)
		if err := c.wsConn.WriteMessage(websocket.TextMessage, []byte(msg)); err != nil {
			return err
		}
		c.subs[sym] = struct{}{}
	}
	return nil
}

// Close stream
func (c *Client) Close() error {
	c.wsMu.Lock()
	defer c.wsMu.Unlock()
	if c.wsConn != nil {
		_ = c.wsConn.Close()
		c.wsConn = nil
	}
	return nil
}

func (c *Client) getConn() *websocket.Conn {
	c.wsMu.Lock()
	defer c.wsMu.Unlock()
	return c.wsConn
}

func (c *Client) ensureConn(ctx context.Context) error {
	var dialErr error
	c.start.Do(func() {
		url := fmt.Sprintf("wss://ws.finnhub.io?token=%s", c.Token)
		conn, _, err := websocket.DefaultDialer.Dial(url, nil)
		if err != nil {
			dialErr = err
			return
		}
		c.wsMu.Lock()
		c.wsConn = conn
		c.wsMu.Unlock()
		go c.readLoop(ctx)
	})
	if dialErr != nil {
		return dialErr
	}
	// If start.Do already ran before, make sure conn still exists; if it was closed, re-dial.
	if c.getConn() == nil {
		// try to re-establish (simple variant: re-run start)
		c.start = sync.Once{} // reset the once
		return c.ensureConn(ctx)
	}
	return nil
}

// --- REST support ---

type quoteResp struct {
	C float64 `json:"c"` // current/last price
	T int64   `json:"t"` // unix seconds
}

func (c *Client) fetchSampleREST(ctx context.Context, asset t.Asset) (t.Sample, error) {
	url := fmt.Sprintf("https://finnhub.io/api/v1/quote?symbol=%s&token=%s", asset.Symbol, c.Token)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return t.Sample{}, err
	}

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return t.Sample{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return t.Sample{}, fmt.Errorf("finnhub: http %d", resp.StatusCode)
	}

	var qr quoteResp
	if err := json.NewDecoder(resp.Body).Decode(&qr); err != nil {
		return t.Sample{}, err
	}
	return t.NewSample(asset, time.Unix(qr.T, 0), qr.C, 0), nil
}

// --- WS stream support ---

type wsTradeMsg struct {
	Type string `json:"type"`
	Data []struct {
		S string  `json:"s"` // symbol
		P float64 `json:"p"` // price
		T int64   `json:"t"` // ms since epoch
		V float64 `json:"v"` // volume
	} `json:"data"`
}

func (c *Client) readLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			_ = c.Close()
			return
		default:
			c.wsMu.Lock()
			conn := c.wsConn
			c.wsMu.Unlock()
			if conn == nil {
				return
			}
			_, msg, err := conn.ReadMessage()
			if err != nil {
				_ = c.Close()
				return // simple: exit (reconnect policy can be added later)
			}
			var m wsTradeMsg
			if err := json.Unmarshal(msg, &m); err != nil || m.Type != "trade" {
				continue
			}
			now := time.Now()
			for _, d := range m.Data {
				sm := t.NewSample(
					t.NewAsset(d.S, cfg.Exchange, cfg.AssetType),
					time.Unix(0, d.T*int64(time.Millisecond)),
					d.P,
					d.V,
				)
				c.setCached(sm, now)
			}
		}
	}
}

// --- Cache helpers ---

type cached struct {
	sample t.Sample
	when   time.Time
}

func (c *Client) getCached(symbol string) (cached, bool) {
	c.cacheMu.RLock()
	defer c.cacheMu.RUnlock()
	s, ok := c.latest[symbol]
	if !ok {
		return cached{}, false
	}
	return cached{sample: s, when: c.stamp[symbol]}, true
}

func (c *Client) setCached(s t.Sample, now time.Time) {
	c.cacheMu.Lock()
	c.latest[s.Asset.Symbol] = s
	c.stamp[s.Asset.Symbol] = now
	c.cacheMu.Unlock()
}
