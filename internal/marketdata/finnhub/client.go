package finnhub

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/gorilla/websocket"

	cfg "github.com/joshskilla/trading-bot/internal/config"
	// md "github.com/joshskilla/trading-bot/internal/marketdata"
	t "github.com/joshskilla/trading-bot/internal/types"
)

// Compile-time check to see if Client implements SampleProvider
// var _ md.SampleProvider = (*Client)(nil)

// finnhub.Client is a Finnhub adapter
type Client struct {
	Token string

	// WS internals (the stream)
	wsMu   sync.Mutex
	wsConn *websocket.Conn
	start  sync.Once

	// Subscriptions (symbols)
	subs map[string]struct{}

	// Latest CLOSED bar per symbol
	barMu       sync.RWMutex
	latestBar   map[string]t.Bar // by symbol
	barInterval time.Duration

	// Bar aggregation
	agg   *t.Aggregator // use with mutex
	aggMu sync.Mutex    // readloop + getLatestBar sync

	// Latest sample
	sampleMu     sync.RWMutex
	latestSample map[string]t.Sample // by symbol

	// Dont hold locks together but if needed, aggMu then barMu then sampleMu
}

func NewClient(token string, interval time.Duration) *Client {
	if interval <= 0 {
		interval = time.Minute
	}
	return &Client{
		Token:        token,
		subs:         make(map[string]struct{}),
		latestSample: make(map[string]t.Sample),
		barInterval:  interval,
		agg:          t.NewAggregator(interval),
		latestBar:    make(map[string]t.Bar),
	}
}

func NewClientWithAssets(ctx context.Context, token string, interval time.Duration, assets []t.Asset) (*Client, error) {
	cl := NewClient(token, interval)
	if err := cl.AddToStream(ctx, assets); err != nil {
		return nil, err
	}
	return cl, nil
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

// Close stream (WS connection)
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

// ensureConn connects the WS if not already connected.
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
		c.start = sync.Once{} // reset so conn can be restarted
		return c.ensureConn(ctx)
	}
	return nil
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

// goroutine: read messages, parse, and update latest sample & bars
// Usually uses trades to update the building bar in aggregator,
// but can also patch the last closed bar if a late trade arrives.
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
			for _, d := range m.Data {
				asset := t.NewAsset(d.S, cfg.Exchange, cfg.AssetType)
				ts := time.Unix(0, d.T*int64(time.Millisecond))

				// Update latest trade
				c.setLatestSample(t.NewSample(asset, ts, d.P, d.V))

				// Updates latestBar or new bar being built
				c.applyTradeToBar(asset, ts, d.P, d.V)
			}
		}
	}
}

// applyTradeToBar routes a trade to the correct place under locks:
// - if same interval as current building bar → update building bar
// - if late for the latest closed bar → patch last closed (latest) bar
// - if late for interval before last closed bar → ignore (too old)
// - if future (ie next) interval → let aggregator roll & close, then start new bar
func (c *Client) applyTradeToBar(asset t.Asset, ts time.Time, price, size float64) {
	tradeIntervalStart := t.IntervalStart(ts, c.barInterval)

	// Guard current building state with aggMu
	c.aggMu.Lock()
	buildingBar, hasBB := c.agg.Curr[asset]

	if hasBB {
		switch {
		case tradeIntervalStart.Equal(buildingBar.Start):
			// Same Interval as Building Bar: update it
			buildingBar.UpdateWithTrade(ts, price, size)
			c.agg.Curr[asset] = buildingBar
			c.aggMu.Unlock()
			return

		case tradeIntervalStart.Before(buildingBar.Start):
			// Earlier Interval than BB: update Last closed bar if possible
			c.aggMu.Unlock()
			_ = c.patchLastClosedBar(asset, ts, price, size)
			return

		default:
			// Future (next) interval: let aggregator close BB & start new bar
			closed := c.agg.PushTrade(asset, ts, price, size)
			c.aggMu.Unlock()
			if closed != nil {
				closed.Status = t.BarStatusAggregated
				c.setLatestBar(*closed)
			}
			return
		}
	}

	// No building bar exists — can only update last closed bar or start new bar
	c.aggMu.Unlock()

	last, ok := c.peekLastClosedBar(asset.Symbol)
	switch {
	case ok && tradeIntervalStart.Before(last.Start):
		// Earlier than last closed bar: ignore
		return
	case ok && tradeIntervalStart.Before(last.End):
		// Same interval as last closed bar: patch it
		_ = c.patchLastClosedBar(asset, ts, price, size)
		return
	default:
		// After last closed bar or no closed bar yet: start/continue via aggregator
		// GetLatest may have lazily deleted building bar or first trade for asset
		// Will create new building bar or continue existing one
		c.aggMu.Lock()
		closed := c.agg.PushTrade(asset, ts, price, size)
		c.aggMu.Unlock()
		if closed != nil {
			closed.Status = t.BarStatusAggregated
			c.setLatestBar(*closed)
		}
		return
	}
}

// --- Latest CLOSED bar ---

// Set the latest closed bar for a symbol
func (c *Client) setLatestBar(b t.Bar) {
	c.barMu.Lock()
	c.latestBar[b.Asset.Symbol] = b
	c.barMu.Unlock()
}

// Peek at the latest closed bar for a symbol
// For client internal use only
func (c *Client) peekLastClosedBar(symbol string) (t.Bar, bool) {
	c.barMu.RLock()
	b, ok := c.latestBar[symbol]
	c.barMu.RUnlock()
	return b, ok
}

// GetLatestBar returns the most recent closed bar for a symbol.
// If we have crossed interval boundaries with no trades, it will use
// zero-volume carry-forward bars up to now, using the last known price.
// For users of client to ensure correctness.
func (c *Client) GetLatestBar(ctx context.Context, asset t.Asset) (t.Bar, bool) {
	// Ensure the asset is added to the stream if not already there
	c.wsMu.Lock()
	_, exists := c.subs[asset.Symbol]
	c.wsMu.Unlock()
	if !exists {
		_ = c.AddToStream(ctx, []t.Asset{asset})
	}

	now := time.Now().UTC()

	// Finalise building bar if required
	_, _ = c.finalizeBuildingIfElapsed(asset, now)

	// Then check for a bar (this only ever returns *closed* bars)
	return c.ensureBarsUpToNow(asset.Symbol, now)
}

// Will ensure bars are up-to-date to now, filling in missed last close interval if necessary with
// a zero-volume carry-forward bar.
func (c *Client) ensureBarsUpToNow(symbol string, now time.Time) (t.Bar, bool) {
	c.barMu.Lock()
	defer c.barMu.Unlock()

	last, has := c.latestBar[symbol]
	currStart := t.IntervalStart(now, c.barInterval) // current interval (building bar interval)
	validLastStart := currStart.Add(-c.barInterval)       // the should be last closed bucket start
	validLastEnd := currStart                             // last closed bucket end

	// If there is a. valid last closed bar return it
    if has && last.End.Equal(currStart) {
        return last, true
    }

	// Get or create sample with last price
	sm, ok := c.getLatestSample(symbol)
	if !ok && !has {
		// no bars and no samples
		return t.Bar{}, false
	}
    if !ok && has {
        // fabricate a 0-volume "sample" using last bar close price
        sm = t.NewSample(last.Asset, validLastStart, last.Close, 0)
    }

	// Synthesize exactly the *last closed* bar
	b := t.NewCarryForwardBarFromSample(sm, validLastStart, validLastEnd)
	c.latestBar[symbol] = b
	return b, true
}

// Close, set & return the current building bar its interval has completed
func (c *Client) finalizeBuildingIfElapsed(asset t.Asset, now time.Time) (t.Bar, bool) {
	c.aggMu.Lock()
	b, ok := c.agg.Curr[asset]
	if !ok || now.Before(b.End) {
		// Building bar is incomplete - still within interval
		c.aggMu.Unlock()
		return t.Bar{}, false
	}
	// Building bar should be complete - close it
	b.Status = t.BarStatusAggregated
	delete(c.agg.Curr, asset) // interval flipped
	c.aggMu.Unlock()

	c.setLatestBar(b)
	return b, true
}

// patchLastClosedBar updates the latest closed bar if ts is in its [Start, End) interval.
func (c *Client) patchLastClosedBar(asset t.Asset, ts time.Time, price, size float64) bool {
	sym := asset.Symbol
	c.barMu.Lock()
	b, ok := c.latestBar[sym]
	if !ok || !b.InRange(ts) {
		c.barMu.Unlock()
		return false
	}
	// Use the helper to update OHLCV/Notional/TradeCount
	b.UpdateWithTrade(ts, price, size)
	c.latestBar[sym] = b
	c.barMu.Unlock()
	return true
}

// --- Latest sample ---

func (c *Client) getLatestSample(symbol string) (t.Sample, bool) {
	c.sampleMu.RLock()
	defer c.sampleMu.RUnlock()
	sm, ok := c.latestSample[symbol]
	return sm, ok
}

func (c *Client) setLatestSample(s t.Sample) {
	c.sampleMu.Lock()
	c.latestSample[s.Asset.Symbol] = s
	c.sampleMu.Unlock()
}
