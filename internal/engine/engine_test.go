package engine

import (
	"fmt"
	"os"
	"testing"
	"time"

	_ "github.com/joshskilla/trading-bot/internal/config"
	"github.com/stretchr/testify/require"
	ds "github.com/joshskilla/trading-bot/internal/datastore"
	"github.com/joshskilla/trading-bot/internal/types"
)

func TestSaveEmptyPortfolio(t *testing.T) {
	// Create a new portfolio & save it
	portfolio := NewPortfolio("UnitTest1EmptyPortfolio", 1000)
	t.Logf("\nBOT_PATH: %s", os.Getenv("BOT_PATH"))
	err := portfolio.SaveToJSON()

	// Assert that there is no error from saving
	require.NoError(t, err)

	path := ds.AbsolutePath(fmt.Sprintf(PortfolioFilePath, portfolio.Name))

	// Assert that the portfolio file exists
	require.True(t, ds.FileExists(path), "Expected portfolio file to exist")

	// Clean up
	os.Remove(path)
}
func TestSaveAndLoadEmptyPortfolio(t *testing.T) {
	// Create a new portfolio & save it
	portfolio := NewPortfolio("UnitTest2EmptyPortfolio", 1000)
	err := portfolio.SaveToJSON()

	// Assert that there is no error
	require.NoError(t, err)

	// Load the portfolio from JSON
	loadedPortfolio, err := LoadPortfolioFromJSON(portfolio.Name)
	require.NoError(t, err)

	// Assert that the loaded portfolio is equal to the original
	require.Equal(t, portfolio, loadedPortfolio)

	path := ds.AbsolutePath(fmt.Sprintf(PortfolioFilePath, portfolio.Name))

	// Clean up
	os.Remove(path)
}

func TestSaveAndLoadPortfolioWithPositions(t *testing.T) {
	// Create a new portfolio & save it
	portfolio := NewPortfolio("UnitTest3PortfolioWithPositions", 1000)
	portfolio.Positions[types.NewAsset("AAPL", "NASDAQ", "stock")] = 10
	portfolio.Positions[types.NewAsset("GOOGL", "NASDAQ", "stock")] = 5
	err := portfolio.SaveToJSON()

	// Assert that there is no error from saving
	require.NoError(t, err)

	// Load the portfolio from JSON
	loadedPortfolio, err := LoadPortfolioFromJSON(portfolio.Name)
	require.NoError(t, err)

	// Assert that the loaded portfolio is equal to the original
	require.Equal(t, portfolio, loadedPortfolio)

	path := ds.AbsolutePath(fmt.Sprintf(PortfolioFilePath, portfolio.Name))

	// Clean up
	os.Remove(path)
}

func TestFlushingOrdersFromPortfolio(t *testing.T) {
	// Create a new portfolio & save it
	portfolio := NewPortfolio("UnitTest4FlushingOrders", 1000)
	portfolio.Positions[types.NewAsset("AAPL", "NASDAQ", "stock")] = 1
	portfolio.Positions[types.NewAsset("GOOGL", "NASDAQ", "stock")] = 1

	// Record some orders
	portfolio.ExecutionHistory = append(portfolio.ExecutionHistory, ExecutionRecord{
		Time:   time.Now(),
		Asset:  types.NewAsset("AAPL", "NASDAQ", "stock"),
		Action: types.Buy,
		Qty:    1,
		Price:  150,
		Cash:   850,
	})
	portfolio.ExecutionHistory = append(portfolio.ExecutionHistory, ExecutionRecord{
		Time:   time.Now(),
		Asset:  types.NewAsset("GOOGL", "NASDAQ", "stock"),
		Action: types.Buy,
		Qty:    1,
		Price:  100,
		Cash:   750,
	})

	// Flush orders to CSV
	err := portfolio.FlushOrdersToFile()
	require.NoError(t, err)
	require.Empty(t, portfolio.ExecutionHistory, "Expected execution history to be cleared after flushing")

	path := ds.AbsolutePath(fmt.Sprintf(OrdersFilePath, portfolio.Name))

	// Assert that the orders file exists
	require.True(t, ds.FileExists(path), "Expected orders file to exist")

	// Clean up
	os.Remove(path)
}

func TestFlushingPositionsFromPortfolio(t *testing.T) {
	// Create a new portfolio & save it
	portfolio := NewPortfolio("UnitTest5FlushingPositions", 1000)
	portfolio.Positions[types.NewAsset("AAPL", "NASDAQ", "stock")] = 1
	portfolio.Positions[types.NewAsset("GOOGL", "NASDAQ", "stock")] = 1

	// Flush positions to CSV
	err := portfolio.FlushPositionsToFile()
	require.NoError(t, err)

	path := ds.AbsolutePath(fmt.Sprintf(PositionsFilePath, portfolio.Name))

	// Assert that the positions file exists
	require.True(t, ds.FileExists(path), "Expected positions file to exist")

	// Clean up
	os.Remove(path)
}