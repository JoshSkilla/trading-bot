package strategy

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"

	ds "github.com/joshskilla/trading-bot/internal/datastore"
	t "github.com/joshskilla/trading-bot/internal/types"
)

const (
	CheckpointFilePath = "data/checkpoints/%s.json"
)

type Checkpoint struct {
	ID         string         `json:"id"`
	Attributes map[string]any `json:"attributes"`
}

type checkpointJSON struct {
	ID     string            `json:"id"`
	Values map[string]any    `json:"values,omitempty"`
	Types  map[string]string `json:"types,omitempty"`
}

func NewCheckpoint(id string, attributes map[string]any) *Checkpoint {
	cp := &Checkpoint{
		ID:         id,
		Attributes: attributes,
	}
	fmt.Printf("Created checkpoint: %+v\n", cp)
	return cp
}

// Load checkpoint from data/checkpoints/<name>.json
func LoadCheckpointFromJSON(name string) (*Checkpoint, error) {
	path := ds.AbsolutePath(fmt.Sprintf(CheckpointFilePath, name))
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// UseNumber so numbers can be handled as json.Number by parsers.
	var cj checkpointJSON
	dec := json.NewDecoder(strings.NewReader(string(raw)))
	dec.UseNumber()
	if err := dec.Decode(&cj); err != nil {
		return nil, err
	}

	attrs := make(map[string]any, len(cj.Values))
	for k, v := range cj.Values {
		typ := ""
		if cj.Types != nil {
			typ = cj.Types[k]
		}
		if val, err := rehydrate(typ, v); err == nil {
			attrs[k] = val
		} else {
			attrs[k] = v // fallback: keep raw
		}
	}

	return &Checkpoint{
		ID:         cj.ID,
		Attributes: attrs,
	}, nil
}

// Marshal checkpoint to JSON and write to data/checkpoints/<id>.json
func (cp *Checkpoint) SaveToJSON() error {
	cj := checkpointJSON{
		ID:     cp.ID,
		Values: make(map[string]any),
		Types:  make(map[string]string),
	}

	for k, v := range cp.Attributes {
		cj.Values[k] = v
		cj.Types[k] = typeTagOf(v)
	}

	data, err := json.MarshalIndent(cj, "", "  ")
	if err != nil {
		return err
	}
	path := ds.AbsolutePath(fmt.Sprintf(CheckpointFilePath, cp.ID))
	return os.WriteFile(path, data, 0644)
}

// Returns a tag of the data's type for parserRegistry lookup
// - builtins: "int", "float64", "bool", "string", ...
// - named structs: "full/pkg/path.TypeName"
func typeTagOf(v any) string {
	rt := reflect.TypeOf(v)
	if rt == nil {
		return "nil"
	}
	if rt.PkgPath() == "" {
		return rt.String() // builtins & composites without pkg path
	}
	if rt.Name() != "" {
		return rt.PkgPath() + "." + rt.Name()
	}
	return rt.String() // fallback
}

// Map type tag -> function that converts a JSON-decoded value into the target Go type
var parserRegistry = map[string]func(any) (any, error){
	// --- Builtin Go types ---
	"int": func(x any) (any, error) {
		switch n := x.(type) {
		case float64:
			return int(n), nil
		case json.Number:
			i, err := n.Int64()
			return int(i), err
		default:
			return nil, fmt.Errorf("cannot parse %T as int", x)
		}
	},
	"int64": func(x any) (any, error) {
		switch n := x.(type) {
		case float64:
			return int64(n), nil
		case json.Number:
			return n.Int64()
		default:
			return nil, fmt.Errorf("cannot parse %T as int64", x)
		}
	},
	"uint": func(x any) (any, error) {
		switch n := x.(type) {
		case float64:
			return uint(n), nil
		case json.Number:
			i, err := n.Int64()
			return uint(i), err
		default:
			return nil, fmt.Errorf("cannot parse %T as uint", x)
		}
	},
	"uint64": func(x any) (any, error) {
		switch n := x.(type) {
		case float64:
			return uint64(n), nil
		case json.Number:
			i, err := n.Int64()
			return uint64(i), err
		default:
			return nil, fmt.Errorf("cannot parse %T as uint64", x)
		}
	},
	"float64": func(x any) (any, error) {
		switch f := x.(type) {
		case float64:
			return f, nil
		case json.Number:
			return f.Float64()
		default:
			return nil, fmt.Errorf("cannot parse %T as float64", x)
		}
	},
	"bool": func(x any) (any, error) {
		b, ok := x.(bool)
		if !ok {
			return nil, fmt.Errorf("cannot parse %T as bool", x)
		}
		return b, nil
	},
	"string": func(x any) (any, error) {
		s, ok := x.(string)
		if !ok {
			return nil, fmt.Errorf("cannot parse %T as string", x)
		}
		return s, nil
	},

	// --- Custom Structs ---
	"github.com/joshskilla/trading-bot/internal/types.Asset": func(x any) (any, error) {
		m, ok := x.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("cannot parse %T as Asset", x)
		}
		var a t.Asset
		if v, ok := m["exchange"].(string); ok {
			a.Exchange = v
		}
		if v, ok := m["symbol"].(string); ok {
			a.Symbol = v
		}
		if v, ok := m["type"].(string); ok {
			a.Type = v
		}
		return a, nil
	},
}

// Applies parser if tag exists; otherwise returns raw value unchanged
func rehydrate(typ string, val any) (any, error) {
	if typ == "" {
		return val, nil
	}
	if fn, ok := parserRegistry[typ]; ok {
		return fn(val)
	}
	return val, nil // unknown tag -> leave value unchanged
}

// Parse command-line args of the form --key value into a map
func ParseKV(args []string) (map[string]string, error) {
	m := make(map[string]string)

	for i := 0; i < len(args)-1; i += 2 {
		if !strings.HasPrefix(args[i], "--") {
			return nil, fmt.Errorf("invalid key %q: must start with --", args[i])
		}

		key := strings.TrimPrefix(args[i], "--")
		val := args[i+1]

		m[key] = val
	}

	if len(args)%2 == 1 {
		return nil, fmt.Errorf("dangling key %q without value", args[len(args)-1])
	}

	return m, nil
}