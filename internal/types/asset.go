package types

type Asset struct {
    Symbol string `json:"symbol"`
    Market string `json:"market"`
    Type   string `json:"type"`
}

func (a Asset) String() string {
    return a.Symbol
}

func NewAsset(symbol, market, assetType string) Asset {
    return Asset{
        Symbol: symbol,
        Market: market,
        Type:   assetType,
    }
}