package engine

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/joshskilla/trading-bot/internal/marketdata/finnhub"
	"github.com/joshskilla/trading-bot/internal/marketdata/alpaca"

	md "github.com/joshskilla/trading-bot/internal/marketdata"
	t "github.com/joshskilla/trading-bot/internal/types"
)

type Trader interface {
	Execute(*Portfolio, t.Signal) (ExecutionRecord, bool)
	FetchSample(ctx context.Context, asset t.Asset) (t.Sample, error)
	Close() error // ensure streams/sessions are cleaned up, ensure idempotency
}

// ----------- LIVE TRADER -----------
type LiveTrader struct {
	Provider md.SampleProvider
}

func NewLiveTrader(ctx context.Context, assets []t.Asset) *LiveTrader {
	cl := finnhub.NewClient(os.Getenv("FINNHUB_API_KEY"))
	// lazily start + add assets
	if err := cl.AddToStream(ctx, assets); err != nil {
		fmt.Println("LiveTrader: stream not enabled, REST only:", err)
	}
	return &LiveTrader{Provider: cl}
}

func (lt *LiveTrader) FetchSample(ctx context.Context, asset t.Asset) (t.Sample, error) {
	return lt.Provider.FetchSample(ctx, asset)
}

func (lt *LiveTrader) Execute(p *Portfolio, sig t.Signal) (ExecutionRecord, bool) {
	// Placeholder - will need to interface with broker API to place live orders
	return ExecutionRecord{}, false
}

func (lt *LiveTrader) Close() error { return lt.Provider.Close() }

// ----------- PAPER TRADER -----------
type PaperTrader struct {
	Provider md.SampleProvider
}

func NewPaperTrader(ctx context.Context, assets []t.Asset) *PaperTrader {
	cl := finnhub.NewClient(os.Getenv("FINNHUB_API_KEY"))
	if err := cl.AddToStream(ctx, assets); err != nil {
		fmt.Println("PaperTrader: stream not enabled, REST only:", err)
	}
	return &PaperTrader{Provider: cl}
}
func (pt *PaperTrader) FetchSample(ctx context.Context, a t.Asset) (t.Sample, error) {
	return pt.Provider.FetchSample(ctx, a)
}

func (pt *PaperTrader) Execute(p *Portfolio, sig t.Signal) (ExecutionRecord, bool) {
	// Placeholder - will need to interface with broker API to place paper orders
	return ExecutionRecord{}, false
}

func (pt *PaperTrader) Close() error { return pt.Provider.Close() }

// ----------- TEST TRADER -----------
type TestTrader struct{
	Provider md.BarProvider
	interval time.Duration
	start    time.Time // inclusive, UTC
	end      time.Time // exclusive, UTC

    cache    map[t.Asset]map[time.Time]t.Bar // asset -> start time -> bar
}

func NewTestTrader(interval time.Duration, start, end time.Time) *TestTrader {
	prov := alpaca.NewClient(os.Getenv("ALPACA_API_KEY"), os.Getenv("ALPACA_API_SECRET"))
	return &TestTrader{
		Provider: prov,
		interval: interval,
		start:    start.UTC(),
		end:      end.UTC(),
		cache:    make(map[t.Asset]map[time.Time]t.Bar),
	}
}

// Preload fetches all bars for each asset in [start, end) and caches them.
func (tt *TestTrader) Preload(ctx context.Context, assets []t.Asset) error {
	s := AlignUTC(tt.start, tt.interval)
	e := tt.end.Add(tt.interval) // pad end (end-exclusive guard)

	for _, a := range assets {
		if err := tt.preloadAsset(ctx, a, s, e); err != nil {
			return err
		}
	}
	return nil
}

func (tt *TestTrader) preloadAsset(ctx context.Context, a t.Asset, s, e time.Time) error {
	bars, err := tt.Provider.FetchBars(ctx, a, s, e, tt.interval)
	if err != nil {
		return err
	}
	if _, ok := tt.cache[a]; !ok {
		tt.cache[a] = make(map[time.Time]t.Bar, len(bars))
	}
	for _, b := range bars {
		tt.cache[a][b.Start.UTC()] = b
	}
	return nil
}

// FetchBarAt returns the bar from cache
// If the asset isn't cached yet, it preloads [start, end) for that asset first.
func (tt *TestTrader) FetchBarAt(ctx context.Context, asset t.Asset, ts time.Time) (t.Bar, bool, error) {
	aligned := AlignUTC(ts.UTC(), tt.interval)

	// If we already have this asset cached, fast path
	if m, ok := tt.cache[asset]; ok {
		if b, ok := m[aligned]; ok {
			return b, true, nil
		}
		// Asset is cached for full window; missing bucket => no data
		return t.Bar{}, false, nil
	}

	// Asset not cached yet: preload full window for this asset
	s := AlignUTC(tt.start, tt.interval)
	e := tt.end.Add(tt.interval)
	if err := tt.preloadAsset(ctx, asset, s, e); err != nil {
		return t.Bar{}, false, err
	}

	// Serve after preload
	if b, ok := tt.cache[asset][aligned]; ok {
		return b, true, nil
	}
	return t.Bar{}, false, nil
}

// Normalises interval boundaries in UTC
func AlignUTC(ts time.Time, interval time.Duration) time.Time {
	sec := int64(interval.Seconds())
	u := ts.Unix()
	return time.Unix((u/int64(sec))*int64(sec), 0).UTC()
}


func (tt *TestTrader) FetchSample(ctx context.Context, asset t.Asset) (t.Sample, error) {
	return t.Sample{}, fmt.Errorf("TestTrader: FetchSample not implemented")
}

func (tt *TestTrader) Execute(p *Portfolio, sig t.Signal) (ExecutionRecord, bool) {
	qty := sig.Qty
	price := sig.Bar.Close
	asset := sig.Bar.Asset
	// live execute will only add it to execHistory once order fulfilled - and get price then
	cost := qty * price
	if sig.Action == t.Buy && p.Cash >= cost {
		p.Cash -= cost
		p.Positions[asset] += qty
		exec := ExecutionRecord{time.Now(), asset, t.Buy, qty, price, p.Cash}
		return exec, true
	}
	return ExecutionRecord{}, false
}

func (tt *TestTrader) Close() error { return nil }
