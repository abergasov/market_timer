package pricer

import (
	"github.com/abergasov/market_timer/internal/config"
	"github.com/abergasov/market_timer/internal/entities"
	"github.com/abergasov/market_timer/internal/logger"
	"github.com/abergasov/market_timer/internal/repository/price"
	"github.com/abergasov/market_timer/internal/service/pricer/etherium"
	"github.com/abergasov/market_timer/internal/service/stopper"
	"github.com/abergasov/market_timer/internal/storage/database"
	"go.uber.org/zap"
)

type serviceConfig struct {
	rpcURL        string
	blockDuration float64
}

func InitObservers(appLog logger.AppLogger, appConf *config.AppConfig, dbConn database.DBConnector) map[string]Observer {
	result := make(map[string]Observer)
	for _, chain := range entities.SupportedChains {
		appLog.Info("init repository", zap.String("chain", chain))
		repo, err := price.InitRepo(dbConn, chain)
		if err != nil {
			appLog.Fatal("unable to init repositories", err, zap.String("chain", chain))
		}
		appLog.Info("init service", zap.String("chain", chain))
		lg := appLog.With(zap.String("chain", chain))
		conf := serviceConfig{
			blockDuration: 1,
		}
		if chain == entities.OPTIMISM {
			continue // for optimish better look at eth price, cause fees depends on eth gas price
		}
		switch chain {
		case entities.ETHEREUM:
			conf = serviceConfig{
				rpcURL:        appConf.ETHRPC,
				blockDuration: 12,
			}
		case entities.ARBITRUM:
			conf.rpcURL = appConf.ARBRPC
		case entities.POLYGON:
			conf.rpcURL = appConf.MATICRPC
		case entities.OPTIMISM:
			conf.rpcURL = appConf.OPRPC
		}
		service := etherium.InitService(lg, repo, chain, conf.rpcURL, conf.blockDuration)
		result[chain] = service
		if err = service.Start(); err != nil {
			appLog.Fatal("unable to start service", err, zap.String("chain", chain))
		}
		stopper.AddStopper(service)
	}
	return result
}
