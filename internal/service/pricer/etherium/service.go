package etherium

import (
	"container/list"
	"context"
	"fmt"
	"sync"

	"github.com/abergasov/market_timer/internal/entities"
	"github.com/abergasov/market_timer/internal/logger"
	"github.com/abergasov/market_timer/internal/repository/price"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

// Service observes gas price for given chain.
// Handle historical data and notify consumers about price changes
type Service struct {
	log    logger.AppLogger
	repo   *price.Repo
	subs   map[float64]chan entities.GasRates
	subsMU sync.RWMutex
	rpcURL string

	ethClient *ethclient.Client

	// graceful shutdown
	ctx    context.Context // context for graceful shutdown
	cancel context.CancelFunc

	headerMU sync.RWMutex
	header   *types.Header

	// gas list
	gasList   *list.List
	gasListMU sync.RWMutex

	// conf
	blockConf BlockSetup
}

func InitService(log logger.AppLogger, repoTimer *price.Repo, rpcURL string, blockDuration float64) *Service {
	return &Service{
		rpcURL:  rpcURL,
		repo:    repoTimer,
		log:     log,
		gasList: list.New(),
		subs:    make(map[float64]chan entities.GasRates),
		blockConf: BlockSetup{
			BlockDuration:          blockDuration,
			MaxKeepBlocks:          uint64(3 * ((24 * 60 * 60) / blockDuration)),
			BlockDownloadBatchSize: 1024,
		},
	}
}

func (s *Service) Start() error {
	for i := 0; i < 10; i++ {
		if err := s.start(); err != nil {
			s.log.Error("unable to start service", err)
			continue
		}
		return nil
	}
	return fmt.Errorf("unable to start service")
}

func (s *Service) start() error {
	s.ctx, s.cancel = context.WithCancel(context.Background())
	s.log.Info("starting service")
	client, err := ethclient.DialContext(s.ctx, s.rpcURL)
	if err != nil {
		return fmt.Errorf("unable to connect to ethereum node: %w", err)
	}
	s.ethClient = client
	go s.observeGas()
	if err = s.DownloadMissedHistory(); err != nil {
		s.log.Fatal("unable to download missed history", err)
	}
	if err = s.prepareGasData(); err != nil {
		s.log.Fatal("unable to prepare gas data", err)
	}
	return nil
}

func (s *Service) Stop() {
	s.log.Info("stopping service")
	s.cancel()
}

func (s *Service) Subscribe(price float64) (chan entities.GasRates, error) {
	s.subsMU.Lock()
	defer s.subsMU.Unlock()
	if _, ok := s.subs[price]; ok {
		return nil, fmt.Errorf("already subscribed")
	}
	ch := make(chan entities.GasRates, 100)
	s.subs[price] = ch
	return ch, nil
}

func (s *Service) Unsubscribe(price float64) {
	s.subsMU.Lock()
	defer s.subsMU.Unlock()
	delete(s.subs, price)
}
