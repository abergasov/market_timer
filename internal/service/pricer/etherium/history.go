package etherium

import (
	"fmt"
	"math/big"

	"go.uber.org/zap"

	"github.com/abergasov/market_timer/internal/entities"
)

type BlockSetup struct {
	BlockDuration          float64 // approximate block duration in seconds
	MaxKeepBlocks          uint64  // how many days in past store gas info
	BlockDownloadBatchSize uint64  // how many blocks download in one iteration
}

func (s *Service) DownloadMissedHistory() error {
	currentBlock, err := s.ethClient.BlockNumber(s.ctx)
	if err != nil {
		return fmt.Errorf("unable to get current block number: %w", err)
	}
	from := currentBlock - s.blockConf.MaxKeepBlocks
	if err = s.repo.DeleteBlocksBefore(big.NewInt(int64(from))); err != nil {
		return fmt.Errorf("unable to delete blocks before %d: %w", from, err)
	}
	if s.chain != entities.ETHEREUM {
		return nil
	}
	gasData, err := s.repo.GetGasData(from, currentBlock)
	if err != nil {
		return fmt.Errorf("unable to get gas data between blocks %d - %d: %w", from, currentBlock, err)
	}
	missedBlocks := s.getMissedBlocks(gasData, from, currentBlock)
	s.log.Info("start download missing blocks...")
	for i, chunk := range missedBlocks {
		s.log.Info("iteration of missing blocks", zap.Int("iteration", i), zap.Int("total", len(missedBlocks)))
		if err = s.downloadAndStoreBlockChunk(chunk); err != nil {
			return fmt.Errorf("unable to download and store block chunk: %w", err)
		}
	}
	s.log.Info("downloading of missed blocks finished")
	return nil
}

func (s *Service) getMissedBlocks(blocks []entities.GasData, from, to uint64) [][]int64 {
	result := make([][]int64, 0, len(blocks)/int(s.blockConf.BlockDownloadBatchSize))
	blocksMap := make(map[uint64]struct{}, len(blocks))
	for _, block := range blocks {
		blocksMap[block.BlockID] = struct{}{}
	}
	for i := from; i <= to; i++ {
		if _, ok := blocksMap[i]; ok {
			continue
		}
		if len(result) == 0 || len(result[len(result)-1]) == int(s.blockConf.BlockDownloadBatchSize) {
			result = append(result, make([]int64, 0, s.blockConf.BlockDownloadBatchSize))
		}
		result[len(result)-1] = append(result[len(result)-1], int64(i))
	}
	return result
}

func (s *Service) downloadAndStoreBlockChunk(chunk []int64) error {
	lastBlock := chunk[len(chunk)-1]
	history, err := s.ethClient.FeeHistory(s.ctx, s.blockConf.BlockDownloadBatchSize, big.NewInt(lastBlock), []float64{25, 50, 75})
	if err != nil {
		return fmt.Errorf("unable to get fee history for last block %d: %w", lastBlock, err)
	}
	var gasData []entities.GasData
	oldestBlock := history.OldestBlock.Uint64()
	for i, block := range history.BaseFee {
		gasData = append(gasData, entities.GasData{
			BlockID: oldestBlock + uint64(i),
			BaseFee: block,
		})
	}
	return s.repo.AddGasData(gasData)
}
