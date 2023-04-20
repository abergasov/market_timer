package etherium

import (
	"fmt"
	"math/big"

	"github.com/abergasov/market_timer/internal/entities"
	"github.com/abergasov/market_timer/internal/utils"
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

// getGasPricePosition loop over all blocks and calculate position of current block
func (s *Service) getGasPricePosition(fee *big.Int) float64 {
	blockHasPriceMore := 0
	s.gasListMU.RLock()
	defer s.gasListMU.RUnlock()
	totalBlocks := s.gasList.Len()
	for el := s.gasList.Front(); el != nil; el = el.Next() {
		if fee.Cmp(el.Value.(entities.GasData).BaseFee) < 0 { // if current block has MORE fees than block in list
			blockHasPriceMore++
		}
	}
	return utils.ToFixed(float64(blockHasPriceMore*100)/float64(totalBlocks), 2)
}
