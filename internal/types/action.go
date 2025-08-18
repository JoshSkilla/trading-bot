package types

// Action represents a trading decision (enum-like)
type Action int

const (
	Unknown Action = iota
	Buy
	Sell
	Hold
)

func (a Action) String() string {
	switch a {
	case Buy:
		return "BUY"
	case Sell:
		return "SELL"
	case Hold:
		return "HOLD"
	default:
		return "UNKNOWN"
	}
}

// Optional: for parsing from strings
func ParseAction(s string) Action {
	switch s {
	case "BUY":
		return Buy
	case "SELL":
		return Sell
	case "HOLD":
		return Hold
	default:
		return Unknown
	}
}
