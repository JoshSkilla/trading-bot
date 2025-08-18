package engine

import (
	"fmt"
	"time"

	ds "github.com/joshskilla/trading-bot/internal/datastore"
	t "github.com/joshskilla/trading-bot/internal/types"
)

type Portfolio struct {
	Name             string
	Cash             float64
	Holdings         map[t.Asset]float64
	ExecutionHistory []ExecutionRecord
	Writer           ds.Writer
}

type ExecutionRecord struct {
	Time   time.Time
	Asset  t.Asset
	Action t.Action
	Qty    float64
	Price  float64
}

func NewPortfolio(name string, cash float64) *Portfolio {
	fileName := fmt.Sprintf("%s_log", name)
	filePath := "results"
	fileType := "csv"
	return &Portfolio{
		Name:             name,
		Cash:             cash,
		Holdings:         make(map[t.Asset]float64),
		ExecutionHistory: []ExecutionRecord{},
		Writer: ds.NewCSVWriter(ds.File{
			Name: fileName,
			Path: filePath,
			Type: fileType,
		}, []string{"Time", "Asset", "Action", "Qty", "Price", "PortfolioValue", "Cash"}),
	}
}

func (p *Portfolio) FlushToCSV(append bool) error {
	// Write p.ExecutionHistory to CSV at path
	// Handle append/overwrite
	// Format: Time,Asset,Action,Qty,Price
	// maybe hold a file interface which has a flush method instead of this
	return nil
}
