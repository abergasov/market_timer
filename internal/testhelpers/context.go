package testhelpers

import (
	"testing"

	"github.com/abergasov/market_timer/internal/logger"
	"github.com/abergasov/market_timer/internal/storage/database"
	"github.com/stretchr/testify/require"
)

func GetTestContext(t *testing.T) database.DBConnector {
	log, err := logger.NewAppLogger("")
	require.NoError(t, err)
	conn, err := database.InitDBConnect(log, "storage.db") //InitMemory(log)
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, conn.Client().Close())
	})
	return conn
}
