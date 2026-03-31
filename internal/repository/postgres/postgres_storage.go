package postgres

import (
	"context"
	"database/sql"
)

type PostgresStorage struct {
	db *sql.DB
}

func New(db *sql.DB) *PostgresStorage {
	return &PostgresStorage{db: db}
}

func (p *PostgresStorage) Ping(ctx context.Context) error {
	return p.db.PingContext(ctx)
}

func (p *PostgresStorage) UpdateGauge(name string, value float64) error {
	return nil
}

func (p *PostgresStorage) GetGauge(name string) (float64, error) {
	return 0, nil
}

func (p *PostgresStorage) GetListGauges() map[string]float64 {
	return map[string]float64{}
}

func (p *PostgresStorage) UpdateCounter(name string, value int64) error {
	return nil
}

func (p *PostgresStorage) GetCounter(name string) (int64, error) {
	return 0, nil
}

func (p *PostgresStorage) GetListCounters() map[string]int64 {
	return map[string]int64{}
}
