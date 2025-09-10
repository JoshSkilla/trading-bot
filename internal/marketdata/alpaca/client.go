// internal/marketdata/alpaca/client.go
package alpaca

import (
	"context"
	"fmt"
	"time"

	alpacaMD "github.com/alpacahq/alpaca-trade-api-go/v3/marketdata"
	cfg "github.com/joshskilla/trading-bot/internal/config"
	md "github.com/joshskilla/trading-bot/internal/marketdata"
	t "github.com/joshskilla/trading-bot/internal/types"
)

// Ensure *Client implements marketdata.BarProvider
var _ md.BarProvider = (*Client)(nil)

type Client struct {
	api *alpacaMD.Client

	barInterval time.Duration
	start       time.Time // inclusive, UTC
	end         time.Time // exclusive, UTC

	cache map[t.Asset]map[time.Time]t.Bar // asset -> start time -> bar
}

// NewClient builds an Alpaca market data client.
func NewClient(apiKey, apiSecret string, barInterval time.Duration, start, end time.Time) *Client {
	opts := alpacaMD.ClientOpts{
		APIKey:    apiKey,
		APISecret: apiSecret,
		// Feed: alpacaAPI.IEX, // default feed is IEX
	}
	return &Client{
		api:         alpacaMD.NewClient(opts),
		barInterval: barInterval,
		start:       start,
		end:         end,
		cache:       make(map[t.Asset]map[time.Time]t.Bar),
	}
}

func (c *Client) Preload(ctx context.Context, assets []t.Asset) error {
	s := t.IntervalStart(c.start, c.barInterval)
	e := c.end.Add(c.barInterval) // pad end (end-exclusive guard)

	for _, a := range assets {
		if err := c.preloadAsset(ctx, a, s, e); err != nil {
			return err
		}
	}
	return nil
}

func (c *Client) preloadAsset(ctx context.Context, a t.Asset, s, e time.Time) error {
	bars, err := c.FetchBars(ctx, a, s, e, c.barInterval)
	if err != nil {
		return err
	}
	if _, ok := c.cache[a]; !ok {
		c.cache[a] = make(map[time.Time]t.Bar, len(bars))
	}
	for _, b := range bars {
		c.cache[a][b.Start.UTC()] = b
	}
	return nil
}

// FetchBars implements BarProvider with explicit start/end and bar interval.
func (client *Client) FetchBars(
	ctx context.Context,
	asset t.Asset,
	start, end time.Time,
	interval time.Duration,
) ([]t.Bar, error) {

	if start.IsZero() || end.IsZero() || !end.After(start) {
		return nil, fmt.Errorf("alpaca: invalid time window (start=%v end=%v)", start, end)
	}

	tf, err := timeFrameFromDuration(interval)
	if err != nil {
		return nil, err
	}

	feed, err := StringToFeed(asset.Exchange)
	if err != nil {
		defaultFeed, err2 := StringToFeed(cfg.Feed)
		if err2 != nil {
			return nil, fmt.Errorf("alpaca: invalid default feed %q: %w", cfg.Feed, err2)
		}
		feed = defaultFeed
	}

	req := alpacaMD.GetBarsRequest{
		TimeFrame:  tf,
		Start:      start.UTC(),
		End:        end.UTC(),
		Adjustment: alpacaMD.Split,
		Feed:       feed,
	}

	bars, err := client.api.GetBars(asset.Symbol, req)
	if err != nil {
		return nil, fmt.Errorf("alpaca GetBars: %w", err)
	}

	var out []t.Bar
	for _, b := range bars {
		out = append(out, t.Bar{
			Asset:    asset,
			Start:    b.Timestamp.UTC(),
			End:      b.Timestamp.UTC().Add(interval),
			Interval: interval,
			Open:     b.Open,
			High:     b.High,
			Low:      b.Low,
			Close:    b.Close,
			Volume:   float64(b.Volume),
		})
	}
	return out, nil
}

func timeFrameFromDuration(d time.Duration) (alpacaMD.TimeFrame, error) {
	switch {
	case d%time.Minute == 0 && d < time.Hour:
		n := int(d / time.Minute)
		return alpacaMD.NewTimeFrame(n, alpacaMD.Min), nil
	case d%time.Hour == 0 && d < 24*time.Hour:
		n := int(d / time.Hour)
		return alpacaMD.NewTimeFrame(n, alpacaMD.Hour), nil
	case d%(24*time.Hour) == 0:
		n := int(d / (24 * time.Hour))
		return alpacaMD.NewTimeFrame(n, alpacaMD.Day), nil
	default:
		return alpacaMD.TimeFrame{}, fmt.Errorf("alpaca: unsupported interval %s (use whole minutes/hours/days)", d)
	}
}

func StringToFeed(s string) (alpacaMD.Feed, error) {
	switch s {
	case "IEX":
		return alpacaMD.IEX, nil
	case "SIP":
		return alpacaMD.SIP, nil
	case "OTC":
		return alpacaMD.OTC, nil
	default:
		return "", fmt.Errorf("alpaca: unknown feed %q", s)
	}
}

// --- BarProvider interface ---

func (c *Client) FetchBarAt(ctx context.Context, asset t.Asset, now time.Time) (t.Bar, bool, error) {
	aligned := t.IntervalStart(now.UTC(), c.barInterval)

	// If we already have this asset cached, fast path
	if m, ok := c.cache[asset]; ok {
		if b, ok := m[aligned]; ok {
			return b, true, nil
		}
		// Asset is cached for full window; missing bucket => no data
		return t.Bar{}, false, nil
	}

	// Asset not cached yet: preload full window for this asset
	s := t.IntervalStart(c.start, c.barInterval)
	e := c.end.Add(c.barInterval)
	if err := c.preloadAsset(ctx, asset, s, e); err != nil {
		return t.Bar{}, false, err
	}

	// Serve after preload
	if b, ok := c.cache[asset][aligned]; ok {
		return b, true, nil
	}
	return t.Bar{}, false, nil
}

func (c *Client) IncludeAssets(ctx context.Context, assets []t.Asset) error {
	return c.Preload(ctx, assets)
}

func (client *Client) Close() error { return nil }
