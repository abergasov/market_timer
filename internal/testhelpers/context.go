package testhelpers

import (
	"fmt"
	"log"
	"testing"

	"github.com/abergasov/market_timer/internal/config"
	"github.com/abergasov/market_timer/internal/logger"
	"github.com/abergasov/market_timer/internal/routes"
	"github.com/abergasov/market_timer/internal/service/notifier"
	"github.com/abergasov/market_timer/internal/service/pricer"
	"github.com/abergasov/market_timer/internal/storage/database"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func GetTestContext(t *testing.T) database.DBConnector {
	appLog, err := logger.NewAppLogger("")
	require.NoError(t, err)
	conn, err := database.InitDBConnect(appLog, "storage.db")
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, conn.Client().Close())
	})
	return conn
}

func SpawnWebServer(t *testing.T, confFile string, dbConn database.DBConnector) {
	appLog, err := logger.NewAppLogger("")
	if err != nil {
		log.Fatalf("unable to create logger: %s", err)
	}
	appConf, err := config.InitConf(confFile)
	require.NoError(t, err, "unable to init config")

	appLog.Info("init repositories")

	serviceNotifier := notifier.NewService(appLog.With(zap.String("service", "notifier")), pricer.InitObservers(appLog, appConf, dbConn))

	appLog.Info("init http service")
	appHTTPServer := routes.InitAppRouter(appLog, serviceNotifier, fmt.Sprintf(":%d", appConf.AppPort))

	t.Cleanup(func() {
		appHTTPServer.Stop()
	})
	go func() {
		if err = appHTTPServer.Run(); err != nil {
			appLog.Fatal("unable to start http service", err)
		}
	}()
}
