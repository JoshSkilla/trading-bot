package engine

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "github.com/joshskilla/trading-bot/config"
	ds "github.com/joshskilla/trading-bot/internal/datastore"
	t "github.com/joshskilla/trading-bot/internal/types"
)

const (
	PortfolioFilePath = "data/portfolios/%s.json"
	ResultsFileDir    = "results"
	ResultsFileType   = "csv"
	OrdersFileName    = "%s_orders"
	PositionsFileName = "%s_positions"
	OrdersFilePath    = "results/%s_orders.csv"
	PositionsFilePath = "results/%s_positions.csv"
)

type Portfolio struct {
	Name             string              `json:"name"`
	Cash             float64             `json:"cash"`
	Positions        map[t.Asset]float64 `json:"positions"`
	ExecutionHistory []ExecutionRecord   `json:"-"`
	OrderWriter      ds.Writer           `json:"-"`
	PositionWriter   ds.Writer           `json:"-"`
}
type portfolioJSON struct {
	Name     string             `json:"name"`
	Cash     float64            `json:"cash"`
	Positions map[string]float64 `json:"positions"`
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
			Dir:  ResultsFileDir,
			Type: ResultsFileType,
		}, []string{"Time", "Asset", "Action", "Qty", "Price", "Cash"}),
		PositionWriter: ds.NewCSVWriter(ds.File{
			Name: fmt.Sprintf(PositionsFileName, name),
			Dir:  ResultsFileDir,
			Type: ResultsFileType,
		}, []string{"Time", "Asset", "Qty", "Price"}),
	}
}

// Marshal portfolio to JSON and write to data/portfolios/<name>.json
func (p *Portfolio) SaveToJSON() error {
	ho := make(map[string]float64, len(p.Positions))
	for a, v := range p.Positions {
		ho[a.String()] = v // define Asset.String() to return a stable key (e.g., "NASDAQ:AAPL")
	}
	data, err := json.MarshalIndent(portfolioJSON{
		Name:     p.Name,
		Cash:     p.Cash,
		Positions: ho,
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
	basePath := os.Getenv("BOT_PATH")
	path := filepath.Join(basePath, fmt.Sprintf(PortfolioFilePath, name))
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var pJ portfolioJSON
	var p Portfolio
	err = json.Unmarshal(data, &pJ)
	if err != nil {
		return nil, err
	}
	p = *NewPortfolio(pJ.Name, pJ.Cash)
	for k, v := range pJ.Positions {
		p.Positions[t.AssetFromString(k)] = v
	}
	return &p, nil
}

func (p *Portfolio) FlushOrdersToFile() error {
	if err := p.OrderWriter.Write(p.ExecutionHistory); err != nil {
		return err
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
		slice := []PositionRecord{record}
		if err := p.PositionWriter.Write(slice); err != nil {
			return err
		}
	}
	return nil
}
