package engine

import (
	"context"
	"fmt"
	"os"
	"time"

	finnhub "github.com/joshskilla/trading-bot/internal/marketdata/finnhub"

	md "github.com/joshskilla/trading-bot/internal/marketdata"
	t "github.com/joshskilla/trading-bot/internal/types"
)

type Trader interface {
	Execute(*Portfolio, t.Signal) (ExecutionRecord, bool)
	FetchSample(ctx context.Context, asset t.Asset) (t.Sample, error)
	Close() // ensure streams/sessions are cleaned up, ensure idempotency
}

// ----------- LIVE TRADER -----------
type LiveTrader struct {
	Provider    md.SampleProvider
	closeStream func()
}

func NewLiveTrader(ctx context.Context, assets []t.Asset) *LiveTrader {
	client := finnhub.NewClient(os.Getenv("FINNHUB_API_KEY"))
	closeStreamFn, err := client.EnableStream(ctx, assets)
	if err != nil {
		fmt.Println("LiveTrader: stream not enabled, using REST only:", err)
	}
	return &LiveTrader{
		Provider:    client,
		closeStream: closeStreamFn,
	}
}

func (lt *LiveTrader) FetchSample(ctx context.Context, asset t.Asset) (t.Sample, error) {
	return lt.Provider.FetchSample(ctx, asset)
}

func (lt *LiveTrader) Execute(p *Portfolio, sig t.Signal) (ExecutionRecord, bool) {
	// Placeholder - will need to interface with broker API to place live orders
	return ExecutionRecord{}, false
}

func (lt *LiveTrader) Close() {
	if lt.closeStream != nil {
		lt.closeStream()
	}
}

// ----------- PAPER TRADER -----------
type PaperTrader struct {
	Provider    md.SampleProvider
	closeStream func()
}

func NewPaperTrader(ctx context.Context, assets []t.Asset) *PaperTrader {
	client := finnhub.NewClient(os.Getenv("FINNHUB_API_KEY"))
	closeStreamFn, err := client.EnableStream(ctx, assets)
	if err != nil {
		fmt.Println("PaperTrader: stream not enabled, using REST only:", err)
	}
	return &PaperTrader{
		Provider:    client,
		closeStream: closeStreamFn,
	}
}
func (pt *PaperTrader) FetchSample(ctx context.Context, asset t.Asset) (t.Sample, error) {
	return pt.Provider.FetchSample(ctx, asset)
}

func (pt *PaperTrader) Execute(p *Portfolio, sig t.Signal) (ExecutionRecord, bool) {
	// Placeholder - will need to interface with broker API to place paper orders
	return ExecutionRecord{}, false
}

func (pt *PaperTrader) Close() {
	if pt.closeStream != nil {
		pt.closeStream()
	}
}

// ----------- TEST TRADER -----------
type TestTrader struct{}

func NewTestTrader() *TestTrader {
	return &TestTrader{}
}

func (tt *TestTrader) FetchSample(ctx context.Context, asset t.Asset) (t.Sample, error) {
	// Placeholder (dummy sample) until I interface with market data provider
	// Will use alpaca
	return t.NewSample(asset, time.Now(), 100, 1), nil
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

func (tt *TestTrader) Close() { /* no-op */ }
