package marketdata

import (
	"context"

	t "github.com/joshskilla/trading-bot/internal/types"
)

type SampleProvider interface {
	// Returns a market observation for the given asset.
	// Will serve from cache if stream enabled and fresh else it'll fall back to REST.
	FetchSample(ctx context.Context, asset t.Asset) (t.Sample, error)

	// EnableStream (optional) starts a background stream that populates a cache.
	EnableStream(ctx context.Context, assets []t.Asset) (stop func(), err error)
}