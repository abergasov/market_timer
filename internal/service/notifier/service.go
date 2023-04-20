package notifier

import (
	"fmt"
	"sync"

	"github.com/abergasov/market_timer/internal/entities"
	"github.com/abergasov/market_timer/internal/logger"
	"github.com/abergasov/market_timer/internal/service/pricer"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Service keep info about price subscriptions
// and notify consumers about price changes
type Service struct {
	log            logger.AppLogger
	priceObservers map[string]pricer.Observer
	priceSubsMU    sync.RWMutex
	priceSubs      map[string]map[float64]map[uuid.UUID]chan entities.GasRates
	proxy          map[string]map[float64]chan entities.GasRates
}

func NewService(log logger.AppLogger, priceObservers map[string]pricer.Observer) *Service {
	subs := make(map[string]map[float64]map[uuid.UUID]chan entities.GasRates)
	proxy := make(map[string]map[float64]chan entities.GasRates)
	for chain := range priceObservers {
		subs[chain] = make(map[float64]map[uuid.UUID]chan entities.GasRates)
		proxy[chain] = make(map[float64]chan entities.GasRates)
	}
	return &Service{
		log:            log,
		priceObservers: priceObservers,
		priceSubs:      subs,
		proxy:          proxy,
	}
}

func (s *Service) NewSubscribe(chain string, percentage float64) (chan entities.GasRates, uuid.UUID, error) {
	s.log.Info("new subscription", zap.String("chain", chain), zap.Float64("percentage", percentage))
	ch := make(chan entities.GasRates, 1_000)
	s.priceSubsMU.Lock()
	defer s.priceSubsMU.Unlock()
	if _, ok := s.priceSubs[chain]; !ok {
		return nil, uuid.Nil, fmt.Errorf("unsupported chain")
	}
	if _, ok := s.priceSubs[chain][percentage]; !ok {
		s.priceSubs[chain][percentage] = make(map[uuid.UUID]chan entities.GasRates, 10)
	}
	if _, ok := s.proxy[chain][percentage]; !ok {
		proxyChan, err := s.priceObservers[chain].Subscribe(percentage)
		if err != nil {
			return nil, uuid.Nil, fmt.Errorf("failed to subscribe to pricer: %w", err)
		}
		s.proxy[chain][percentage] = proxyChan
		go s.processPriceChange(chain, percentage, proxyChan)
	}
	chanID := uuid.New()
	s.priceSubs[chain][percentage][chanID] = ch
	return ch, chanID, nil
}

func (s *Service) UnSubscribe(chain string, percentage float64, chanID uuid.UUID) {
	s.log.Info("unsubscribe", zap.String("chain", chain), zap.Float64("percentage", percentage), zap.String("chanID", chanID.String()))
	s.priceSubsMU.Lock()
	delete(s.priceSubs[chain][percentage], chanID)
	subs := len(s.priceSubs[chain][percentage])
	s.priceSubsMU.Unlock()
	if subs > 0 {
		s.priceObservers[chain].Unsubscribe(percentage)
	}
}

func (s *Service) processPriceChange(chain string, percentage float64, source chan entities.GasRates) {
	for rates := range source {
		s.priceSubsMU.RLock()
		for _, ch := range s.priceSubs[chain][percentage] {
			ch <- rates
		}
		s.priceSubsMU.RUnlock()
	}
}
