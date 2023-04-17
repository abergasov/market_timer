package pricer

import (
	"github.com/abergasov/market_timer/internal/entities"
	"github.com/ethereum/go-ethereum/core/types"
	"go.uber.org/zap"
)

func (s *Service) observeGas() {
	ch := make(chan *types.Header, 100)
	go s.handleHead(ch)
	if _, err := s.ethClient.SubscribeNewHead(s.ctx, ch); err != nil {
		s.log.Fatal("unable to subscribe to new head", err)
	}
}

func (s *Service) handleHead(ch chan *types.Header) {
	for h := range ch {
		s.log.Info("new header", zap.String("block", h.Number.String()), zap.String("base_fee", h.BaseFee.String()))
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
	}
}
