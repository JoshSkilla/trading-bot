package types

import (
	"fmt"
	"strings"
)

type Asset struct {
	Symbol   string `json:"symbol"`
	Exchange string `json:"exchange"`
	Type     string `json:"type"`
}

func (a *Asset) String() string {
	return fmt.Sprintf("%s:%s:%s", a.Symbol, a.Exchange, a.Type)
}

func AssetFromString(s string) Asset {
	parts := strings.Split(s, ":")
	if len(parts) != 3 {
		return Asset{Symbol: "", Exchange: "", Type: ""}
	}
	return Asset{
		Symbol:   parts[0],
		Exchange: parts[1],
		Type:     parts[2],
	}
}

func NewAsset(symbol, exchange, assetType string) Asset {
	return Asset{
		Symbol:   symbol,
		Exchange: exchange,
		Type:     assetType,
	}
}
