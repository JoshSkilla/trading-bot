package datastore

import (
	t "github.com/joshskilla/trading-bot/internal/types"
	"time"
)

type DataStore interface {
	Get(asset string, t time.Time) (t.Bar, bool)
}
/*
	To suppport CSV data, real time data, in memory stores and cached databases
*/