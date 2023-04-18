package entities

import (
	"fmt"
	"math/big"
)

type GasRates struct {
	Percentage float64
	BlockID    uint64
	BaseFee    *big.Int
}

type GasData struct {
	BlockID uint64
	BaseFee *big.Int
}

type GasDataMapper struct {
	BlockID uint64 `db:"block_id"`
	BaseFee string `db:"base_fee"`
}

func (g *GasDataMapper) ToGasData() (GasData, error) {
	baseFee, ok := big.NewInt(0).SetString(g.BaseFee, 10)
	if !ok {
		return GasData{}, fmt.Errorf("unable to parse base fee: %s", g.BaseFee)
	}
	return GasData{
		BlockID: g.BlockID,
		BaseFee: baseFee,
	}, nil
}
