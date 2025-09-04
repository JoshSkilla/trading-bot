package engine

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
	"path/filepath"

	ds "github.com/joshskilla/trading-bot/internal/datastore"
	t "github.com/joshskilla/trading-bot/internal/types"
	_ "github.com/joshskilla/trading-bot/config"
)

const (
	PortfolioFilePath       = "data/portfolios/%s.json"
	ResultsFileDir          = "results"
	ResultsFileType         = "csv"
	OrdersFileName          = "%s_orders"
	PositionsFileName       = "%s_positions"
	OrdersFilePath          = "results/%s_orders.csv"
	PositionsFilePath       = "results/%s_positions.csv"
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
	return &Portfolio{
		Name:             name,
		Cash:             cash,
		Positions:        make(map[t.Asset]float64),
		ExecutionHistory: []ExecutionRecord{},
		OrderWriter: ds.NewCSVWriter(ds.File{
			Name: fmt.Sprintf(OrdersFileName, name),
			Path: ResultsFileDir,
			Type: ResultsFileType,
		}, []string{"Time", "Asset", "Action", "Qty", "Price", "Cash"}),
		PositionWriter: ds.NewCSVWriter(ds.File{
			Name: fmt.Sprintf(PositionsFileName, name),
			Path: ResultsFileDir,
			Type: ResultsFileType,
		}, []string{"Time", "Asset", "Qty", "Price"}),
	}
}

// Marshal portfolio to JSON and write to data/portfolios/<name>.json
func (p *Portfolio) SaveToJSON() error {
    type out struct {
        Name     string             `json:"name"`
        Cash     float64            `json:"cash"`
        Holdings map[string]float64 `json:"holdings"`
    }
	ho := make(map[string]float64, len(p.Positions))
    for a, v := range p.Positions {
        ho[a.String()] = v // define Asset.String() to return a stable key (e.g., "NASDAQ:AAPL")
    }
	data, err := json.MarshalIndent(out{
		Name:     p.Name,
		Cash:     p.Cash,
		Holdings: ho,
	}, "", "  ")
	if err != nil {
		return err
	}
	basePath := os.Getenv("BOT_PATH")
	path := filepath.Join(basePath, fmt.Sprintf(PortfolioFilePath, p.Name))
	return os.WriteFile(path, data, 0644)
}

// Load portfolio from data/portfolios/<name>.json
func LoadPortfolioFromJSON(name string) (*Portfolio, error) {
	path := fmt.Sprintf(PortfolioFilePath, name)
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
	for _, record := range p.ExecutionHistory {
		if err := p.OrderWriter.Write(record); err != nil {
			return err
		}
	}
	p.ExecutionHistory = []ExecutionRecord{}
	return nil
}

func (p *Portfolio) FlushPositionsToFile() error {
	now := time.Now()
	for asset, qty := range p.Positions {
		price := 0.0 // TODO: send requests for price
		record := PositionRecord{
			Time:  now,
			Asset: asset,
			Qty:   qty,
			Price: price,
		}
		if err := p.PositionWriter.Write(record); err != nil {
			return err
		}
	}
	return nil
}
