package strategy

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

const (
	CheckpointFilePath = "data/checkpoints/%s.json"
)

type Checkpoint struct {
	ID         string         `json:"id"`
	Attributes map[string]any `json:"attributes"`
}

func NewCheckpoint(ctx context.Context, id string, attributes map[string]any) (*Checkpoint, error) {
	cp := &Checkpoint{
		ID:         id,
		Attributes: attributes,
	}
	fmt.Printf("Created checkpoint: %+v\n", cp)
	return cp, nil
}

// Load checkpoint from data/checkpoints/<name>.json
func LoadCheckpointFromJSON(name string) (*Checkpoint, error) {
	path := fmt.Sprintf(CheckpointFilePath, name)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cp Checkpoint
	err = json.Unmarshal(data, &cp)
	if err != nil {
		return nil, err
	}
	return &cp, nil
}

// Marshal checkpoint to JSON and write to data/checkpoints/<id>.json
func (cp *Checkpoint) SaveToJSON() error {
	data, err := json.MarshalIndent(cp, "", "  ")
	if err != nil {
		return err
	}
	path := fmt.Sprintf(CheckpointFilePath, cp.ID)
	return os.WriteFile(path, data, 0644)
}

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
