package types

type Asset struct {
    Symbol string
    Market string
    Type   string
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