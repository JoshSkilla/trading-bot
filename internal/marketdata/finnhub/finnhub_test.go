package finnhub

import (
	"context"
	"os"
	"testing"
	"time"

	// md "github.com/joshskilla/trading-bot/internal/marketdata"
	cfg "github.com/joshskilla/trading-bot/config"
	"github.com/joshskilla/trading-bot/internal/types"
)

/***************
 * Integration test (guarded)
 *
 * Requires: FINNHUB_API_KEY in environment.
 * Skips otherwise.
 ***************/

func TestFinnhub_FetchSample_Integration(t *testing.T) {
	token := os.Getenv("FINNHUB_API_KEY")
	if token == "" {
		t.Skip("FINNHUB_API_KEY not set; skipping integration test")
	}

	client := NewClient(token)
	t.Cleanup(func() { _ = client.Close() }) // ensure websocket is closed

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	reqAsset := types.NewAsset("AAPL", cfg.Exchange, cfg.AssetType)
	if err := client.AddToStream(ctx, []types.Asset{reqAsset}); err != nil {
		t.Logf("\n [Finnhub] Stream not enabled, will use REST only: %v \n", err)
	} else {
		t.Logf("\n [Finnhub] Stream subscription sent for %s\n", reqAsset.Symbol)
	}

	// Give the websocket some time to potentially receive a trade from the stream
	time.Sleep(500 * time.Millisecond)

	sampCtx, sampCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer sampCancel()

	s, err := client.FetchSample(sampCtx, reqAsset)
	if err != nil {
		t.Fatalf("FetchSample error: %v", err)
	}
	t.Logf("\n [Finnhub] Fetched sample: %+v \n", s.Pretty())

	if s.Asset.Symbol != reqAsset.Symbol {
		t.Fatalf("expected %s, got %s", reqAsset.Symbol, s.Asset.Symbol)
	}

	// Attribute sanity checks (positive price & reasonable timestamp)
	if s.Price <= 0 {
		t.Fatalf("expected positive price, got %f", s.Price)
	}
	if s.Time.Before(time.Date(2015, 1, 1, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("unexpected old timestamp: %v", s.Time)
	}

	// Idempotent close
	if err := client.Close(); err != nil {
		t.Fatalf("Close failed: %v", err)
	}
	// Second close should also be safe
	if err := client.Close(); err != nil {
		t.Fatalf("second Close failed: %v", err)
	}
}
