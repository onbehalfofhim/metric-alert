package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
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
	CryptoKey        string        `env:"CRYPTO_KEY"`
	ConfigFile       string        `env:"CONFIG"`
}

func defaultServerConfig() ServerConfig {
	return ServerConfig{
		RunAddr:          "localhost:8080",
		StoreIntervalRaw: 300,
		FilePath:         "./metrics.txt",
		Restore:          true,
	}
}

// обработка аргументов командной строки
// и сохраняет их значения в структуре
func ParseServerFlags() (ServerConfig, error) {
	_ = godotenv.Load()

	// проставялем дефолтные значения
	cfg := defaultServerConfig()

	// парсим переданные серверу аргументы командной строки в зарегистрированные переменные
	flags := parseServerFlags()

	// CONFIG можно задать через env
	if flags.ConfigFile == "" {
		flags.ConfigFile = os.Getenv("CONFIG")
	}

	if flags.ConfigFile != "" {
		var jc serverConfigJSON
		data, err := os.ReadFile(flags.ConfigFile)
		if err != nil {
			return cfg, err
		}
		if err := json.Unmarshal(data, &jc); err != nil {
			return cfg, err
		}
		applyJSON(&cfg, jc)
	}

	// применяем явно переданные флаги поверх JSON-файла
	applyFlags(&cfg, flags)

	// парсим переменные окружения
	err := env.Parse(&cfg)
	if err != nil {
		return cfg, fmt.Errorf("can't parse environment variables: %w", err)
	}

	if cfg.StoreIntervalRaw <= 0 {
		return cfg, fmt.Errorf("invalid store interval: %d (must be > 0)", cfg.StoreIntervalRaw)
	}
	cfg.StoreInterval = time.Duration(cfg.StoreIntervalRaw) * time.Second

	return cfg, nil
}

type serverFlags struct {
	RunAddr          string
	StoreIntervalRaw int
	FilePath         string
	Restore          bool
	DatabaseDSN      string
	Key              string
	AuditFile        string
	AuditURL         string
	CryptoKey        string
	ConfigFile       string
}

func parseServerFlags() serverFlags {
	var f serverFlags

	flag.StringVar(&f.RunAddr, "a", "", "address and port to run server")
	flag.StringVar(&f.FilePath, "f", "", "file path to write metrics")
	flag.BoolVar(&f.Restore, "r", false, "load metrics from storage")

	flag.IntVar(&f.StoreIntervalRaw, "i", 0, "store interval in seconds")

	flag.StringVar(&f.DatabaseDSN, "d", "", "address to connect DataBase")

	flag.StringVar(&f.Key, "k", "", "signing key")

	flag.StringVar(&f.AuditFile, "audit-file", "", "file path to write audit")
	flag.StringVar(&f.AuditURL, "audit-url", "", "addres to send audit")

	flag.StringVar(&f.CryptoKey, "crypto-key", "", "file path to private key for encryption")

	flag.StringVar(&f.ConfigFile, "c", "", "path to config file")
	flag.StringVar(&f.ConfigFile, "config", "", "path to config file")

	flag.Parse()

	return f
}

type serverConfigJSON struct {
	RunAddr          *string `json:"address"`
	StoreIntervalRaw *int    `json:"store_interval"`
	FilePath         *string `json:"store_file"`
	Restore          *bool   `json:"restore"`
	DatabaseDSN      *string `json:"database_dsn"`
	CryptoKey        *string `json:"crypto_key"`
}

func applyJSON(cfg *ServerConfig, jc serverConfigJSON) {
	if jc.RunAddr != nil {
		cfg.RunAddr = *jc.RunAddr
	}
	if jc.StoreIntervalRaw != nil {
		cfg.StoreIntervalRaw = *jc.StoreIntervalRaw
	}
	if jc.FilePath != nil {
		cfg.FilePath = *jc.FilePath
	}
	if jc.Restore != nil {
		cfg.Restore = *jc.Restore
	}
	if jc.DatabaseDSN != nil {
		cfg.DatabaseDSN = *jc.DatabaseDSN
	}
	if jc.CryptoKey != nil {
		cfg.CryptoKey = *jc.CryptoKey
	}
}

func applyFlags(cfg *ServerConfig, f serverFlags) {
	flag.Visit(func(fl *flag.Flag) {
		switch fl.Name {
		case "a":
			cfg.RunAddr = f.RunAddr

		case "i":
			cfg.StoreIntervalRaw = f.StoreIntervalRaw

		case "f":
			cfg.FilePath = f.FilePath

		case "r":
			cfg.Restore = f.Restore

		case "d":
			cfg.DatabaseDSN = f.DatabaseDSN

		case "k":
			cfg.Key = f.Key

		case "audit-file":
			cfg.AuditFile = f.AuditFile

		case "audit-url":
			cfg.AuditURL = f.AuditURL

		case "crypto-key":
			cfg.CryptoKey = f.CryptoKey
		}
	})
}
