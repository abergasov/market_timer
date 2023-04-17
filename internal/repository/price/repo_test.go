package price_test

import (
	"math/big"
	"testing"

	"github.com/abergasov/market_timer/internal/entities"
	"github.com/abergasov/market_timer/internal/repository/price"
	"github.com/abergasov/market_timer/internal/testhelpers"
	"github.com/stretchr/testify/require"
)

func TestRepo(t *testing.T) {
	conn := testhelpers.GetTestContext(t)
	repo, err := price.InitRepo(conn, entities.ETH)
	require.NoError(t, err)

	sampleData := []entities.GasData{
		{
			BlockID: 123,
			BaseFee: big.NewInt(123),
		},
		{
			BlockID: 234,
			BaseFee: big.NewInt(234),
		},
		{
			BlockID: 345,
			BaseFee: big.NewInt(345),
		},
		{
			BlockID: 456,
			BaseFee: big.NewInt(456),
		},
		{
			BlockID: 567,
			BaseFee: big.NewInt(567),
		},
		{
			BlockID: 678,
			BaseFee: big.NewInt(678),
		},
		{
			BlockID: 789,
			BaseFee: big.NewInt(789),
		},
	}
	require.NoError(t, repo.AddGasData(sampleData))
	t.Run("check data", func(t *testing.T) {
		gasData, errC := repo.LoadAllBlocks()
		require.NoError(t, errC)
		require.Equal(t, sampleData, gasData)
	})
	t.Run("should load blocks between range", func(t *testing.T) {
		gasData, errC := repo.GetGasData(sampleData[1].BlockID, sampleData[4].BlockID)
		require.NoError(t, errC)
		require.Equal(t, sampleData[1:5], gasData)
	})
	t.Run("should drop blocks before", func(t *testing.T) {
		require.NoError(t, repo.DeleteBlocksBefore(big.NewInt(345)))
		gasData, errC := repo.LoadAllBlocks()
		require.NoError(t, errC)
		require.Equal(t, sampleData[2:], gasData)
	})
}
