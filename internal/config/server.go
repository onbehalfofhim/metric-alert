package config

import (
	"flag"
	"fmt"
	"time"

	"github.com/caarlos0/env/v11"
)

type ServerConfig struct {
	RunAddr       string        `env:"ADDRESS"`
	StoreInterval time.Duration `env:"STORE_INTERVAL"`
	FilePath      string        `env:"FILE_STORAGE_PATH"`
	Restore       bool          `env:"RESTORE"`
}

// обработка аргументов командной строки
// и сохраняет их значения в структуре
func ParseServerFlags() (ServerConfig, error) {
	var cfg ServerConfig

	// регистрируем переменную RunAddr
	// как аргумент -a со значением по умолчанию
	flag.StringVar(&cfg.RunAddr, "a", "localhost:8080", "address and port to run server")
	flag.StringVar(&cfg.FilePath, "f", "./metrics.txt", "file path to write metrics")
	flag.BoolVar(&cfg.Restore, "r", true, "load metrics from storage")

	var storeInterval int
	flag.IntVar(&storeInterval, "i", 300, "store interval in seconds")

	// парсим переданные серверу аргументы командной строки в зарегистрированные переменные
	flag.Parse()

	if storeInterval <= 0 {
		return cfg, fmt.Errorf("invalid store interval: %d (must be > 0)", storeInterval)
	}
	cfg.StoreInterval = time.Duration(storeInterval) * time.Second

	// парсим переменные окружения
	err := env.Parse(&cfg)
	if err != nil {
		return cfg, fmt.Errorf("can't parse environment variables: %w", err)
	}

	return cfg, nil
}
