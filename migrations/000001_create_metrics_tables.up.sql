-- migrations/000001_create_metrics_tables.up.sql
-- Создание таблицы для хранения метрик с типом gauges
CREATE TABLE IF NOT EXISTS gauges (
    name TEXT PRIMARY KEY,
    value DOUBLE PRECISION
);

-- Создание таблицы для хранения метрик с типом counters
CREATE TABLE IF NOT EXISTS counters (
    name TEXT PRIMARY KEY,
    value BIGINT
);