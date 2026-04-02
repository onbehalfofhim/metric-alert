package postgres

import (
	"context"
	"database/sql"

	"github.com/onbehalfofhim/metric-alert/internal/repository"
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
	query := `INSERT INTO gauges (name, value)
		VALUES ($1, $2)
		ON CONFLICT (name)
		DO UPDATE SET value = $2
	`
	_, err := p.db.Exec(query, name, value)
	if err != nil {
		return err
	}

	return nil
}

func (p *PostgresStorage) GetGauge(name string) (float64, error) {
	query := `SELECT value 
		FROM gauges 
		WHERE name = $1
	`
	row := p.db.QueryRowContext(context.Background(), query, name)

	var value float64
	err := row.Scan(&value)
	if err != nil {
		return 0, repository.ErrMetricNotFound
	}

	return value, nil
}

func (p *PostgresStorage) GetListGauges() map[string]float64 {
	query := `
		SELECT name, value FROM gauges
	`
	rows, err := p.db.Query(query)
	if err != nil {
		return map[string]float64{}
	}
	defer rows.Close()

	res := make(map[string]float64)
	for rows.Next() {
		var k string
		var v float64
		rows.Scan(&k, &v)
		res[k] = v
	}

	if err := rows.Err(); err != nil {
		return map[string]float64{}
	}

	return res
}

func (p *PostgresStorage) UpdateCounter(name string, value int64) error {
	query := `INSERT INTO counters (name, value)
		VALUES ($1, $2)
		ON CONFLICT (name)
		DO UPDATE SET value = counters.value + $2
	`
	_, err := p.db.Exec(query, name, value)
	if err != nil {
		return err
	}

	return nil
}

func (p *PostgresStorage) GetCounter(name string) (int64, error) {
	query := `SELECT value 
		FROM counters 
		WHERE name = $1
	`
	row := p.db.QueryRowContext(context.Background(), query, name)

	var value int64
	err := row.Scan(&value)
	if err != nil {
		return 0, repository.ErrMetricNotFound
	}

	return value, nil
}

func (p *PostgresStorage) GetListCounters() map[string]int64 {
	query := `
		SELECT name, value FROM counters
	`
	rows, err := p.db.Query(query)
	if err != nil {
		return map[string]int64{}
	}
	defer rows.Close()

	res := make(map[string]int64)
	for rows.Next() {
		var k string
		var v int64
		rows.Scan(&k, &v)
		res[k] = v
	}

	if err := rows.Err(); err != nil {
		return map[string]int64{}
	}

	return res
}
