package engine

import (
	"context"
	"os"
	"time"

	"github.com/joshskilla/trading-bot/internal/marketdata/alpaca"
	"github.com/joshskilla/trading-bot/internal/marketdata/finnhub"

	md "github.com/joshskilla/trading-bot/internal/marketdata"
	t "github.com/joshskilla/trading-bot/internal/types"
)

type Trader interface {
	Execute(*Portfolio, t.Signal) (ExecutionRecord, bool)
	FetchBarAt(ctx context.Context, asset t.Asset, ts time.Time) (t.Bar, bool, error)
	IncludeAssets(ctx context.Context, assets []t.Asset) error
	Close() error // ensure streams/sessions are cleaned up, ensure idempotency
}

// ----------- LIVE TRADER -----------
type LiveTrader struct {
	Provider md.BarProvider
}

// Ensure LiveTrader implements Trader
var _ Trader = (*LiveTrader)(nil)

func NewLiveTrader(ctx context.Context, interval time.Duration) *LiveTrader {
	cl := finnhub.NewClient(os.Getenv("FINNHUB_API_KEY"), interval)
	return &LiveTrader{Provider: cl}
}

func (lt *LiveTrader) IncludeAssets(ctx context.Context, assets []t.Asset) error {
	return lt.Provider.IncludeAssets(ctx, assets)
}

func (lt *LiveTrader) FetchBarAt(ctx context.Context, asset t.Asset, ts time.Time) (t.Bar, bool, error) {
	return lt.Provider.FetchBarAt(ctx, asset, ts)
}

func (lt *LiveTrader) Execute(p *Portfolio, sig t.Signal) (ExecutionRecord, bool) {
	// Placeholder - will need to interface with broker API to place live orders
	return ExecutionRecord{}, false
}

func (lt *LiveTrader) Close() error { return lt.Provider.Close() }

// ----------- PAPER TRADER -----------
type PaperTrader struct {
	Provider md.BarProvider
}

// Ensure PaperTrader implements Trader
var _ Trader = (*PaperTrader)(nil)

func NewPaperTrader(ctx context.Context, interval time.Duration) *PaperTrader {
	cl := finnhub.NewClient(os.Getenv("FINNHUB_API_KEY"), interval)
	return &PaperTrader{Provider: cl}
}

func (pt *PaperTrader) FetchBarAt(ctx context.Context, asset t.Asset, ts time.Time) (t.Bar, bool, error) {
	return pt.Provider.FetchBarAt(ctx, asset, ts)
}

func (pt *PaperTrader) IncludeAssets(ctx context.Context, assets []t.Asset) error {
	return pt.Provider.IncludeAssets(ctx, assets)
}

func (pt *PaperTrader) Execute(p *Portfolio, sig t.Signal) (ExecutionRecord, bool) {
	// Placeholder - will need to interface with broker API to place paper orders
	return ExecutionRecord{}, false
}

func (pt *PaperTrader) Close() error { return pt.Provider.Close() }

// ----------- TEST TRADER -----------
type TestTrader struct {
	Provider md.BarProvider
	interval time.Duration
	start    time.Time // inclusive, UTC
	end      time.Time // exclusive, UTC
}

// Ensure TestTrader implements Trader
var _ Trader = (*TestTrader)(nil)

func NewTestTrader(interval time.Duration, start, end time.Time) *TestTrader {
	prov := alpaca.NewClient(os.Getenv("ALPACA_API_KEY"), os.Getenv("ALPACA_API_SECRET"), interval, start, end)
	return &TestTrader{
		Provider: prov,
		interval: interval,
		start:    start.UTC(),
		end:      end.UTC(),
	}
}

func (tt *TestTrader) FetchBarAt(ctx context.Context, asset t.Asset, ts time.Time) (t.Bar, bool, error) {
	return tt.Provider.FetchBarAt(ctx, asset, ts)
}

func (tt *TestTrader) IncludeAssets(ctx context.Context, assets []t.Asset) error {
	return tt.Provider.IncludeAssets(ctx, assets)
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
		exec := ExecutionRecord{time.Now().UTC(), asset, t.Buy, qty, price, p.Cash}
		return exec, true
	}
	return ExecutionRecord{}, false
}

func (tt *TestTrader) Close() error { return nil }
