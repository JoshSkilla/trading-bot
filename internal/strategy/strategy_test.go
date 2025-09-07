package strategy

import (
	"fmt"
	"os"
	"testing"

	_ "github.com/joshskilla/trading-bot/internal/config"
	ds "github.com/joshskilla/trading-bot/internal/datastore"
	"github.com/stretchr/testify/require"
	types "github.com/joshskilla/trading-bot/internal/types"
)

func TestSaveEmptyCheckpoint(t *testing.T) {
	// Create a new checkpoint & save it
	checkpoint := NewCheckpoint("UnitTest1EmptyCheckpoint", map[string]any{})

	err := checkpoint.SaveToJSON()

	// Assert that there is no error from saving
	require.NoError(t, err)

	path := ds.AbsolutePath(fmt.Sprintf(CheckpointFilePath, checkpoint.ID))

	// Assert that the checkpoint file exists
	require.True(t, ds.FileExists(path), "Expected checkpoint file to exist")

	// Clean up
	os.Remove(path)
}

func TestSaveAndLoadEmptyCheckpoint(t *testing.T) {
	// Create a new checkpoint & save it
	checkpoint := NewCheckpoint("UnitTest2EmptyCheckpoint", map[string]any{})
	err := checkpoint.SaveToJSON()
	
	// Assert that there is no error
	require.NoError(t, err)

	// Load the checkpoint from JSON
	loadedCheckpoint, err := LoadCheckpointFromJSON(checkpoint.ID)
	require.NoError(t, err)

	// Assert that the loaded checkpoint matches the original
	require.Equal(t, checkpoint.ID, loadedCheckpoint.ID)
	require.Equal(t, checkpoint.Attributes, loadedCheckpoint.Attributes)

	path := ds.AbsolutePath(fmt.Sprintf(CheckpointFilePath, checkpoint.ID))

	// Clean up
	os.Remove(path)
}

func TestSaveAndLoadCheckpointWithAttributes(t *testing.T) {
	// Create a new checkpoint & save it
	attrs := map[string]any{
		"last_processed_index": 42,
		"some_flag":           true,
		"threshold":           0.75,
		"description":         "Test checkpoint with attributes",
		"asset":               types.NewAsset("AAPL", "NASDAQ", "stock"),
	}
	checkpoint := NewCheckpoint("UnitTest3CheckpointWithAttributes", attrs)
	err := checkpoint.SaveToJSON()

	// Assert that there is no error from saving
	require.NoError(t, err)

	// Load the checkpoint from JSON
	loadedCheckpoint, err := LoadCheckpointFromJSON(checkpoint.ID)
	require.NoError(t, err)

	// Assert that the loaded checkpoint matches the original
	require.Equal(t, checkpoint.ID, loadedCheckpoint.ID)
	require.Equal(t, checkpoint.Attributes, loadedCheckpoint.Attributes)

	path := ds.AbsolutePath(fmt.Sprintf(CheckpointFilePath, checkpoint.ID))

	// Clean up
	os.Remove(path)
}
