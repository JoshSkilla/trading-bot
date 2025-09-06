package marketdata

import (
	"context"

	t "github.com/joshskilla/trading-bot/internal/types"
)

type SampleProvider interface {
	// Returns a market observation for the given asset.
	// Will serve from cache if stream enabled and fresh else it'll fall back to REST.
	FetchSample(ctx context.Context, asset t.Asset) (t.Sample, error)

	// Lazily start stream and/or extend the WS stream with more assets.
	AddToStream(ctx context.Context, assets []t.Asset) error

	// Releases stream/resources (idempotent)
	Close() error                                            
}