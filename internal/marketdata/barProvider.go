package marketdata

import (
	"context"
	"time"

	t "github.com/joshskilla/trading-bot/internal/types"
)

type BarProvider interface {
	// Fetches historical bars of length duration for the given asset and time range.
	// FetchBars(ctx context.Context, asset t.Asset, start, end time.Time, duration time.Duration) ([]t.Bar, error)

	FetchBarAt(ctx context.Context, asset t.Asset, ts time.Time) (t.Bar, bool)

	IncludeAssets(ctx context.Context, assets []t.Asset)

	// Releases stream/resources (idempotent)
	Close() error
}