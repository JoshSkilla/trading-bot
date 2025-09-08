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
}

// NewClient builds an Alpaca market data client.
func NewClient(apiKey, apiSecret string) *Client {
	opts := alpacaMD.ClientOpts{
		APIKey:    apiKey,
		APISecret: apiSecret,
		// Feed: alpacaAPI.IEX, // default feed is IEX
	}
	return &Client{api: alpacaMD.NewClient(opts)}
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

func (client *Client) Close() error { return nil }

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
