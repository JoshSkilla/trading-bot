package engine

import (
	"time"
)

const (
	AssetType              = "stock"
	Exchange               = "IEX"
	ExchangeTimeZone       = "America/New_York"
	Currency               = "USD"
	OpenHour               = 9
	OpenMinute             = 30
	ClosingHour            = 16
	ClosingMinute          = 0
	MaxLiveTradingDuration = 10 * time.Hour
)
