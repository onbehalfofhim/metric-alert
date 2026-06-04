package config

import (
	"flag"
	"fmt"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

type ServerConfig struct {
	RunAddr          string        `env:"ADDRESS"`
	StoreInterval    time.Duration `env:"-"`
	StoreIntervalRaw int           `env:"STORE_INTERVAL"`
	FilePath         string        `env:"FILE_STORAGE_PATH"`
	Restore          bool          `env:"RESTORE"`
	DatabaseDSN      string        `env:"DATABASE_DSN"`
	Key              string        `env:"KEY"`
	AuditFile        string        `env:"AUDIT_FILE"`
	AuditURL         string        `env:"AUDIT_URL"`
}

// обработка аргументов командной строки
// и сохраняет их значения в структуре
func ParseServerFlags() (ServerConfig, error) {
	var cfg ServerConfig

	_ = godotenv.Load()

	// регистрируем переменную RunAddr
	// как аргумент -a со значением по умолчанию
	flag.StringVar(&cfg.RunAddr, "a", "localhost:8080", "address and port to run server")
	flag.StringVar(&cfg.FilePath, "f", "./metrics.txt", "file path to write metrics")
	flag.BoolVar(&cfg.Restore, "r", true, "load metrics from storage")

	flag.IntVar(&cfg.StoreIntervalRaw, "i", 300, "store interval in seconds")

	flag.StringVar(&cfg.DatabaseDSN, "d", "", "address to connect DataBase")

	flag.StringVar(&cfg.Key, "k", "", "signing key")

	flag.StringVar(&cfg.AuditFile, "audit-file", "", "file path to write audit")
	flag.StringVar(&cfg.AuditFile, "audit-url", "", "addres to send audit")

	// парсим переданные серверу аргументы командной строки в зарегистрированные переменные
	flag.Parse()

	if cfg.StoreIntervalRaw <= 0 {
		return cfg, fmt.Errorf("invalid store interval: %d (must be > 0)", cfg.StoreIntervalRaw)
	}

	// парсим переменные окружения
	err := env.Parse(&cfg)
	if err != nil {
		return cfg, fmt.Errorf("can't parse environment variables: %w", err)
	}

	cfg.StoreInterval = time.Duration(cfg.StoreIntervalRaw) * time.Second

	return cfg, nil
}
