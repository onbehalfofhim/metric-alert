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

func (s *PostgresStorage) UpdateGauge(name string, value float64) error {
	return nil
}

func (s *PostgresStorage) GetGauge(name string) (float64, error) {
	return 0, nil
}

func (s *PostgresStorage) GetListGauges() map[string]float64 {
	return map[string]float64{}
}

func (s *PostgresStorage) UpdateCounter(name string, value int64) error {
	return nil
}

func (s *PostgresStorage) GetCounter(name string) (int64, error) {
	return 0, nil
}

func (s *PostgresStorage) GetListCounters() map[string]int64 {
	return map[string]int64{}
}
