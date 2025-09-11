package alpaca

import (
	"context"
	"os"
	"testing"
	"time"

	cfg "github.com/joshskilla/trading-bot/internal/config"
	types "github.com/joshskilla/trading-bot/internal/types"
	"github.com/stretchr/testify/require"
)

/***************
 * Integration test (guarded)
 *
 * Requires: ALPACA_API_KEY & ALPACA_API_SECRET in environment.
 * Skips otherwise.
 ***************/

func TestAlpaca_FetchBars_Integration(t *testing.T) {
	// Skip if no keys in env, so CI doesn’t fail
	apiKey := os.Getenv("ALPACA_API_KEY")
	apiSecret := os.Getenv("ALPACA_API_SECRET")
	if apiKey == "" || apiSecret == "" {
		t.Skip("ALPACA_API_KEY or ALPACA_API_SECRET not set, skipping integration test")
	}

	// One trading week: Jan 25–29, 2021 (UTC)
	// Give ET timezone midnight to midnight but convert to utc
	// Since time before 00:00 ET snaps to 00:00 ET / 05:00 UTC
	loc, err := time.LoadLocation(cfg.ExchangeTimeZone)
	start := time.Date(2021, 1, 25, 0, 0, 0, 0, loc).UTC()
	end := time.Date(2021, 1, 30, 0, 0, 0, 0, loc).UTC()
	interval := 24 * time.Hour

	c := NewClient(apiKey, apiSecret, interval, start, end)
	defer c.Close()

	ctx := context.Background()
	reqAsset := types.NewAsset("GME", cfg.Exchange, cfg.AssetType)

	require.NoError(t, err)

	bars, err := c.FetchBars(ctx, reqAsset, start, end, interval)
	require.NoError(t, err)
	require.NotEmpty(t, bars)

	t.Logf("\n [Alpaca] got %d bars for GME between %v and %v", len(bars), start, end)
	for i, b := range bars[:5] { // print first 5 bars pretty
		t.Logf("\n[Alpaca] bar[%d]: \n %s", i, b.Pretty())
	}

	require.True(t, bars[0].Start.After(start) || bars[0].Start.Equal(start))
	require.True(t, bars[len(bars)-1].Start.Before(end))

	require.True(t, bars[0].Start.Equal(bars[0].Start.Truncate(24*time.Hour).Add(5*time.Hour)))
}
