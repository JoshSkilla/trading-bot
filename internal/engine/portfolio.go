package engine

import (
	"fmt"
	"time"

	ds "github.com/joshskilla/trading-bot/internal/datastore"
	t "github.com/joshskilla/trading-bot/internal/types"
)

type Portfolio struct {
	name             string
	Cash             float64
	Positions         map[t.Asset]float64
	ExecutionHistory []ExecutionRecord
	OrderWriter      ds.Writer
	PositionWriter   ds.Writer
}

type PositionRecord struct {
	Time   time.Time
	Asset  t.Asset
	Qty    float64
	Price  float64
}

type ExecutionRecord struct {
	Time   time.Time
	Asset  t.Asset
	Action t.Action
	Qty    float64
	Price  float64
	Cash   float64
}

func NewPortfolio(name string, cash float64) *Portfolio {
	filePath := "results"
	fileType := "csv"
	return &Portfolio{
		name:             name,
		Cash:             cash,
		Positions:         make(map[t.Asset]float64),
		ExecutionHistory: []ExecutionRecord{},
		OrderWriter: ds.NewCSVWriter(ds.File{
			Name: fmt.Sprintf("%s_orders", name),
			Path: filePath,
			Type: fileType,
		}, []string{"Time", "Asset", "Action", "Qty", "Price", "Cash"}),
		PositionWriter: ds.NewCSVWriter(ds.File{
			Name: fmt.Sprintf("%s_positions", name),
			Path: filePath,
			Type: fileType,
		}, []string{"Time", "Asset", "Qty", "Price"}),
	}
}

func (p *Portfolio) Name() string {
	return p.name
}

func (p *Portfolio) FlushOrdersToFile(append bool) error {
	// Write p.ExecutionHistory to CSV at path
	// Handle append/overwrite
	// Format: Time,Asset,Action,Qty,Price
	// maybe hold a file interface which has a flush method instead of this
	return nil
}

func (p *Portfolio) FlushPositionsToFile(append bool) error {
	// Write p.Positions to CSV at path
	// Handle append/overwrite
	// Format: Time,Asset,Qty,Price
	return nil
}