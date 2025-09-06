package finnhub

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
	"github.com/gorilla/websocket"
	
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

	// internal stream+cache
	wsConn  *websocket.Conn
	cacheMu sync.RWMutex
	latest  map[string]t.Sample  // by symbol
	stamp   map[string]time.Time // last update time
}

func NewClient(token string) *Client {
	return &Client{
		HTTP:         &http.Client{Timeout: 10 * time.Second},
		Token:        token,
		MaxStaleness: 2 * time.Second,
		latest:       make(map[string]t.Sample),
		stamp:        make(map[string]time.Time),
	}
}

// If streaming is enabled and cache is valid/fresh, serves from cache
// Else falls back to REST /quote
func (c *Client) FetchSample(ctx context.Context, asset t.Asset) (t.Sample, error) {
	if c.wsConn != nil {
		if sm, ok := c.getCached(asset.Symbol); ok && time.Since(sm.when) <= c.MaxStaleness {
			return sm.sample, nil
		}
	}
	return c.fetchSampleREST(ctx, asset)
}

// EnableStream starts the WS stream which populates the internal cache
// Call once at startup if you want streaming
func (c *Client) EnableStream(ctx context.Context, assets []t.Asset) (stop func(), err error) {
	url := fmt.Sprintf("wss://ws.finnhub.io?token=%s", c.Token)
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return nil, err
	}
	c.wsConn = conn

	// subscribe to assets
	for _, a := range assets {
		msg := fmt.Sprintf(`{"type":"subscribe","symbol":"%s"}`, a.Symbol)
		if err := c.wsConn.WriteMessage(websocket.TextMessage, []byte(msg)); err != nil {
			_ = c.wsConn.Close()
			c.wsConn = nil
			return nil, err
		}
	}

	go c.readLoop(ctx)

	// Safe to call multiple times, will do nothing if already closed
	stop = func() {
		if c.wsConn != nil {
			_ = c.wsConn.Close()
			c.wsConn = nil
		}
	}
	return stop, nil
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
			return
		default:
			_, msg, err := c.wsConn.ReadMessage()
			if err != nil {
				return // (todo: reconnection/backoff?)
			}
			var m wsTradeMsg
			if err := json.Unmarshal(msg, &m); err != nil {
				continue
			}
			if m.Type != "trade" {
				continue
			}

			now := time.Now()
			for _, d := range m.Data {
				sm := t.NewSample(
					t.NewAsset(d.S, "Finnhub", "stock"),
					time.Unix(0, d.T*int64(time.Millisecond)),
					d.P, d.V,
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
