package types

type Asset struct {
    Symbol  string `json:"symbol"`
    Exchange string `json:"exchange"`
    Type    string `json:"type"`
}

func (a Asset) String() string {
    return a.Symbol
}

func NewAsset(symbol, exchange, assetType string) Asset {
    return Asset{
        Symbol:  symbol,
        Exchange: exchange,
        Type:    assetType,
    }
}