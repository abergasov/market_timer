package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/abergasov/market_timer/internal/config"
	"github.com/abergasov/market_timer/internal/logger"
	"github.com/abergasov/market_timer/internal/routes"
	"github.com/abergasov/market_timer/internal/service/notifier"
	"github.com/abergasov/market_timer/internal/service/pricer"
	"github.com/abergasov/market_timer/internal/service/stopper"
	"github.com/abergasov/market_timer/internal/storage/database"
	"go.uber.org/zap"
)

var (
	dbPath   = "storage.db"
	confFile = "configs/app_conf.yml"
	appHash  = os.Getenv("GIT_HASH")
)

func main() {
	appLog, err := logger.NewAppLogger(appHash)
	if err != nil {
		log.Fatalf("unable to create logger: %s", err)
	}
	appLog.Info("app starting", zap.String("conf", confFile))
	appConf, err := config.InitConf(confFile)
	if err != nil {
		appLog.Fatal("unable to init config", err, zap.String("config", confFile))
	}
	defer stopper.Stop()
	appLog.Info("create storage connections")
	dbConn, err := getDBConnect(appLog, dbPath)
	if err != nil {
		appLog.Fatal("unable to connect to db", err, zap.String("host", dbPath))
	}
	stopper.AddStopper(dbConn)

	serviceNotifier := notifier.NewService(
		appLog.With(zap.String("service", "notifier")),
		pricer.InitObservers(appLog, appConf, dbConn),
	)

	appLog.Info("init http service")
	appHTTPServer := routes.InitAppRouter(appLog, serviceNotifier, fmt.Sprintf(":%d", appConf.AppPort))
	stopper.AddStopper(appHTTPServer)
	go func() {
		if err = appHTTPServer.Run(); err != nil {
			appLog.Fatal("unable to start http service", err)
		}
	}()

	// register app shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGINT, syscall.SIGQUIT, os.Interrupt, syscall.SIGTERM)
	<-c // This blocks the main thread until an interrupt is received
}

func getDBConnect(log logger.AppLogger, dbPath string) (*database.DBConnect, error) {
	for i := 0; i < 5; i++ {
		dbConnect, err := database.InitDBConnect(log.With(zap.String("service", "db")), dbPath)
		if err == nil {
			return dbConnect, nil
		}
		log.Error("can't connect to db", err, zap.Int("attempt", i))
		time.Sleep(time.Duration(i) * time.Second * 5)
	}
	return nil, fmt.Errorf("can't connect to db")
}
