package pricer

import (
	"container/list"
	"context"
	"fmt"
	"sync"

	"github.com/abergasov/market_timer/internal/logger"
	"github.com/abergasov/market_timer/internal/repository/price"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

// Service observes gas price for given chain.
// Handle historical data and notify consumers about price changes
type Service struct {
	log  logger.AppLogger
	repo *price.Repo
	subs map[uint32]struct{}

	ethClient *ethclient.Client

	// graceful shutdown
	ctx    context.Context // context for graceful shutdown
	cancel context.CancelFunc

	headerMU sync.RWMutex
	header   *types.Header

	// gas list
	gasList   *list.List
	gasListMU sync.RWMutex
}

func InitService(log logger.AppLogger, repoTimer *price.Repo, rpcURL string) (*Service, error) {
	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to ethereum node: %w", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	return &Service{
		ctx:       ctx,
		cancel:    cancel,
		ethClient: client,
		repo:      repoTimer,
		log:       log,
		gasList:   list.New(),
	}, nil
}

func (s *Service) Start() {
	s.log.Info("starting service")
	go s.observeGas()
	if err := s.DownloadMissedHistory(); err != nil {
		s.log.Fatal("unable to download missed history", err)
	}
	if err := s.prepareGasData(); err != nil {
		s.log.Fatal("unable to prepare gas data", err)
	}
}

func (s *Service) Stop() {
	s.log.Info("stopping service")
	s.cancel()
}
