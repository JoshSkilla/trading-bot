package engine

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	ds "github.com/joshskilla/trading-bot/internal/datastore"
	t "github.com/joshskilla/trading-bot/internal/types"
)

type Portfolio struct {
	Name             string              `json:"name"`
	Cash             float64             `json:"cash"`
	Positions        map[t.Asset]float64 `json:"positions"`
	ExecutionHistory []ExecutionRecord   `json:"-"`
	OrderWriter      ds.Writer           `json:"-"`
	PositionWriter   ds.Writer           `json:"-"`
}

type PositionRecord struct {
	Time  time.Time
	Asset t.Asset
	Qty   float64
	Price float64
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
		Name:             name,
		Cash:             cash,
		Positions:        make(map[t.Asset]float64),
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

// Marshal portfolio to JSON and write to data/portfolios/<name>.json
func (p *Portfolio) SaveToJSON() error {
	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return err
	}
	path := fmt.Sprintf("data/portfolios/%s.json", p.Name)
	return os.WriteFile(path, data, 0644)
}

// Load portfolio from data/portfolios/<name>.json
func LoadPortfolioFromJSON(name string) (*Portfolio, error) {
	path := fmt.Sprintf("data/portfolios/%s.json", name)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var p Portfolio
	err = json.Unmarshal(data, &p)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (p *Portfolio) FlushOrdersToFile() error {
	// Write p.ExecutionHistory to CSV at path
	// Handle append/overwrite
	// Format: Time,Asset,Action,Qty,Price
	// maybe hold a file interface which has a flush method instead of this
	return nil
}

func (p *Portfolio) FlushPositionsToFile() error {
	// Write p.Positions to CSV at path
	// Handle append/overwrite
	// Format: Time,Asset,Qty,Price
	return nil
}
