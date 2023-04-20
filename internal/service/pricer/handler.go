package pricer

import (
	"github.com/abergasov/market_timer/internal/entities"
	"github.com/ethereum/go-ethereum/core/types"
	"go.uber.org/zap"
)

func (s *Service) observeGas() {
	ch := make(chan *types.Header, 100)
	go s.handleHead(ch)
	sub, err := s.ethClient.SubscribeNewHead(s.ctx, ch)
	if err != nil {
		s.log.Fatal("unable to subscribe to new head", err)
	}
	go func() {
		for subErr := range sub.Err() {
			s.log.Fatal("subscription error", subErr)
		}
	}()
}

func (s *Service) handleHead(ch chan *types.Header) {
	for h := range ch {
		s.headerMU.Lock()
		s.header = h
		s.headerMU.Unlock()
		s.addBlock(h.Number.Uint64(), h.BaseFee)
		if err := s.repo.AddGasData([]entities.GasData{
			{
				BlockID: h.Number.Uint64(),
				BaseFee: h.BaseFee,
			},
		}); err != nil {
			s.log.Error("unable to add gas data", err)
		}
		percent := s.getGasPricePosition(h.BaseFee)
		s.log.Info("blocks has more price than current", zap.Float64("percent", percent), zap.String("block", h.Number.String()), zap.String("base_fee", h.BaseFee.String()))
		s.notifySubs(entities.GasRates{
			BlockID:    h.Number.Uint64(),
			BaseFee:    h.BaseFee,
			Percentage: percent,
		})
	}
}

func (s *Service) notifySubs(rate entities.GasRates) {
	s.subsMU.RLock()
	defer s.subsMU.RUnlock()
	for percent, sub := range s.subs {
		if rate.Percentage <= percent {
			sub <- rate
		}
	}
}
