package engine

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	_ "github.com/joshskilla/trading-bot/config"
	"github.com/stretchr/testify/require"
	ds "github.com/joshskilla/trading-bot/internal/datastore"
)

func TestSaveEmptyPortfolio(t *testing.T) {
	// Create a new portfolio & save it
	portfolio := NewPortfolio("UnitTest1EmptyPortfolio", 1000)
	t.Logf("\nBOT_PATH: %s", os.Getenv("BOT_PATH"))
	err := portfolio.SaveToJSON()

	// Assert that there is no error
	require.NoError(t, err)
	basePath := os.Getenv("BOT_PATH")
	path := filepath.Join(basePath, fmt.Sprintf(PortfolioFilePath, portfolio.Name))

	// Assert that the portfolio file exists
	require.True(t, ds.FileExists(path), "Expected portfolio file to exist")

	// Clean up
	os.Remove(path)
}

func TestSavePortfolioWithPositions(t *testing.T) {

}

func TestSaveAndLoadPortfolio(t *testing.T) {
	
}

