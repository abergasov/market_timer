package price

import (
	"fmt"
	"math/big"
	"strings"

	"github.com/abergasov/market_timer/internal/entities"

	"github.com/abergasov/market_timer/internal/storage/database"
)

type Repo struct {
	conn      database.DBConnector
	tableName string
}

func InitRepo(db database.DBConnector, networkName string) (*Repo, error) {
	repo := &Repo{
		conn:      db,
		tableName: fmt.Sprintf("gas_history_%s", networkName),
	}
	if err := repo.migrate(); err != nil {
		return nil, fmt.Errorf("unable to migrate repo %s: %w", repo.tableName, err)
	}
	return repo, nil
}

func (r *Repo) migrate() error {
	q := []string{
		fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (block_id INTEGER PRIMARY KEY, base_fee TEXT)`, r.tableName),
	}
	for _, query := range q {
		if _, err := r.conn.Client().Exec(query); err != nil {
			return fmt.Errorf("unable to migrate: %w", err)
		}
	}
	return nil
}

// DeleteBlocksBefore deletes all blocks before given block number
func (r *Repo) DeleteBlocksBefore(before *big.Int) error {
	_, err := r.conn.Client().Exec(fmt.Sprintf(`DELETE FROM %s WHERE block_id < ?`, r.tableName), before.String())
	return err
}

// LoadAllBlocks loads all blocks from the database
func (r *Repo) LoadAllBlocks() ([]entities.GasData, error) {
	sql := fmt.Sprintf(`SELECT * FROM %s ORDER BY block_id ASC`, r.tableName)
	rows, err := r.conn.Client().Queryx(sql)
	if err != nil {
		return nil, fmt.Errorf("unable to load all blocks: %w", err)
	}
	defer rows.Close()
	result := make([]entities.GasData, 0, 100_000)
	for rows.Next() {
		var gd entities.GasDataMapper
		if err = rows.StructScan(&gd); err != nil {
			return nil, fmt.Errorf("unable to scan gas data: %w", err)
		}
		data, err := gd.ToGasData()
		if err != nil {
			return nil, fmt.Errorf("unable to convert gas data: %w", err)
		}
		result = append(result, data)
	}
	return result, nil
}

// AddGasData save data about fees into storage for future decisions
func (r *Repo) AddGasData(payload []entities.GasData) error {
	sqlAppend := make([]string, 0)
	sqlParams := make([]interface{}, 0)
	for _, item := range payload {
		sqlAppend = append(sqlAppend, "(?, ?)")
		sqlParams = append(sqlParams, item.BlockID, item.BaseFee.String())
	}
	sql := fmt.Sprintf("INSERT INTO %s (block_id, base_fee) VALUES "+strings.Join(sqlAppend, ",")+` ON CONFLICT(block_id) DO NOTHING`, r.tableName)
	_, err := r.conn.Client().Exec(sql, sqlParams...)
	return err
}

// GetGasData returns data about fees for given block range
func (r *Repo) GetGasData(from, to uint64) ([]entities.GasData, error) {
	sql := fmt.Sprintf(`SELECT * FROM %s WHERE block_id >= ? AND block_id <= ? ORDER BY block_id ASC`, r.tableName)
	rows, err := r.conn.Client().Queryx(sql, from, to)
	if err != nil {
		return nil, fmt.Errorf("unable to get gas data: %w", err)
	}
	defer rows.Close()
	result := make([]entities.GasData, 0, 100_000)
	for rows.Next() {
		var gd entities.GasDataMapper
		if err = rows.StructScan(&gd); err != nil {
			return nil, fmt.Errorf("unable to scan gas data: %w", err)
		}
		data, err := gd.ToGasData()
		if err != nil {
			return nil, fmt.Errorf("unable to convert gas data: %w", err)
		}
		result = append(result, data)
	}
	return result, nil
}
