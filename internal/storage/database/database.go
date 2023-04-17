package database

import (
	"fmt"

	"github.com/abergasov/market_timer/internal/logger"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3" // justifying it
)

type DBConnect struct {
	db  *sqlx.DB
	log logger.AppLogger
}

func InitDBConnect(log logger.AppLogger, dbPath string) (*DBConnect, error) {
	db, err := sqlx.Connect("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("error connect to db: %w", err)
	}

	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("error ping to db: %w", err)
	}
	conn := &DBConnect{
		db:  db,
		log: log,
	}
	return conn, nil
}

func InitMemory(log logger.AppLogger) (DBConnector, error) {
	db, err := sqlx.Connect("sqlite3", ":memory:")
	if err != nil {
		return nil, fmt.Errorf("error connect to db: %w", err)
	}
	return &DBConnect{db: db, log: log}, err
}

func (d *DBConnect) Stop() {
	d.log.Info("close db connection")
	if err := d.db.Close(); err != nil {
		d.log.Error("error close db connection", err)
	}
}

func (d *DBConnect) Close() error {
	return d.db.Close()
}

func (d *DBConnect) Client() *sqlx.DB {
	return d.db
}
