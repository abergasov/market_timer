package pricer

import (
	"fmt"
	"math/big"

	"github.com/abergasov/market_timer/internal/entities"
)

func (s *Service) prepareGasData() error {
	blocks, err := s.repo.LoadAllBlocks()
	if err != nil {
		return fmt.Errorf("unable to load blocks: %w", err)
	}
	s.gasListMU.Lock()
	defer s.gasListMU.Unlock()
	for _, block := range blocks {
		s.gasList.PushFront(block)
	}
	return nil
}

func (s *Service) addBlock(blockID uint64, fees *big.Int) {
	s.gasListMU.Lock()
	defer s.gasListMU.Unlock()
	s.gasList.PushFront(entities.GasData{
		BlockID: blockID,
		BaseFee: fees,
	})
	if uint64(s.gasList.Len()) > MaxKeepBlocks {
		s.gasList.Remove(s.gasList.Back())
	}
}
