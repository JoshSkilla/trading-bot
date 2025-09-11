package finnhub

import (
	"context"
	"os"
	"testing"
	"time"

	cfg "github.com/joshskilla/trading-bot/internal/config"
	"github.com/joshskilla/trading-bot/internal/types"
)

/***************
 * Integration test (guarded)
 *
 * Requires: FINNHUB_API_KEY in environment.
 * Requires: Market to be open (per config trading hours).
 * Skips otherwise.
 ***************/

func TestFinnhub_FetchSample_Integration(t *testing.T) {
	tradingHours := types.TradingHours{
		OpenHour:    cfg.OpenHour,
		OpenMinute:  cfg.OpenMinute,
		CloseHour:   cfg.ClosingHour,
		CloseMinute: cfg.ClosingMinute,
		WeekendsOff: true,
		ExchangeTZ:  cfg.ExchangeTimeZone,
	}

	token := os.Getenv("FINNHUB_API_KEY")
	if token == "" {
		t.Skip("FINNHUB_API_KEY not set; skipping integration test")
	}

	if !tradingHours.IsOpenAt(time.Now().UTC()) {
		t.Skip("Market closed; skipping integration test")
	}

	interval := 5 * time.Second
	client := NewClient(token, interval)
	t.Cleanup(func() { _ = client.Close() }) // ensure websocket is closed

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	reqAsset := types.NewAsset("AAPL", cfg.Exchange, cfg.AssetType)
	if err := client.IncludeAssets(ctx, []types.Asset{reqAsset}); err != nil {
		t.Fatalf("failed to subscribe asset: %v", err)
	} else {
		t.Logf("[Finnhub] Stream subscription sent for %s", reqAsset.Symbol)
	}

	// Give the websocket a moment to receive initial trades (if market is open).
	time.Sleep(2 * time.Second)

	// Poll FetchBarAt for up to ~10s so we don't rely on an exact boundary moment.
	deadline := time.Now().Add(10 * time.Second)
	var (
		bar       types.Bar
		ok        bool
		err       error
		checkedAt time.Time
	)
	for time.Now().Before(deadline) {
		now := time.Now().UTC()
		bar, ok, err = client.FetchBarAt(ctx, reqAsset, now)
		if err != nil {
			t.Fatalf("FetchBarAt error: %v", err)
		}
		if ok {
			checkedAt = now
			break
		}
		time.Sleep(500 * time.Millisecond)
	}

	if !ok {
		t.Logf("[Finnhub] No bar available yet for %s (market closed or no trades)", reqAsset.Symbol)
		return
	}

	// Verify alignment to the last-closed interval end using the same 'now' we queried with.
	expectedEnd := types.IntervalStart(checkedAt, interval)
	if !bar.End.Equal(expectedEnd) {
		t.Errorf("expected bar.End %v, got %v", expectedEnd, bar.End)
	}
	if bar.Interval != interval {
		t.Errorf("expected interval %v, got %v", interval, bar.Interval)
	}

	// Log in your existing style, and clearly include TradeCount.
	t.Logf("[Finnhub] Bar for %s:\n %s [%s - %s] (%s)\n OHLC: %.2f %.2f %.2f %.2f  Volume: %.2f  Trades: %d  Status: %v",
		bar.Asset.Symbol,
		bar.Asset.Symbol,
		bar.Start.UTC().Format("2006-01-02 15:04:05 UTC"),
		bar.End.UTC().Format("2006-01-02 15:04:05 UTC"),
		bar.Interval.String(),
		bar.Open, bar.High, bar.Low, bar.Close,
		bar.Volume,
		bar.TradeCount,
		bar.Status,
	)

	// Idempotent close check
	if err := client.Close(); err != nil {
		t.Fatalf("Close failed: %v", err)
	}
	if err := client.Close(); err != nil {
		t.Fatalf("second Close failed: %v", err)
	}
}
