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
	RunAddr       string        `env:"ADDRESS"`
	StoreInterval time.Duration `env:"STORE_INTERVAL"`
	FilePath      string        `env:"FILE_STORAGE_PATH"`
	Restore       bool          `env:"RESTORE"`
	DatabaseDSN   string        `env:"DATABASE_DSN"`
	Key           string        `env:"KEY"`
	AuditFile     string        `env:"AUDIT_FILE"`
	AuditURL      string        `env:"AUDIT_URL"`
	CryptoKey     string        `env:"CRYPTO_KEY"`
	ConfigFile    string        `env:"CONFIG"`
}

func defaultServerConfig() ServerConfig {
	return ServerConfig{
		RunAddr:       "localhost:8080",
		StoreInterval: 300 * time.Second,
		FilePath:      "./metrics.txt",
		Restore:       true,
	}
}

// обработка аргументов командной строки
// и сохраняет их значения в структуре
func ParseServerFlags() (ServerConfig, error) {
	_ = godotenv.Load()

	// проставялем дефолтные значения
	cfg := defaultServerConfig()

	// парсим переданные серверу аргументы командной строки в зарегистрированные переменные
	fs, flags, err := parseServerFlags()
	if err != nil {
		return cfg, err
	}

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

		if err := applyJSON(&cfg, jc); err != nil {

			return cfg, err
		}
	}

	// применяем явно переданные флаги поверх JSON-файла
	applyFlags(&cfg, fs, flags)

	// парсим переменные окружения
	if err := env.Parse(&cfg); err != nil {
		return cfg, fmt.Errorf("can't parse environment variables: %w", err)
	}

	if cfg.StoreInterval < 0 {
		return cfg, fmt.Errorf("invalid store interval: %s", cfg.StoreInterval)
	}

	return cfg, nil
}

type serverFlags struct {
	RunAddr       string
	StoreInterval time.Duration
	FilePath      string
	Restore       bool
	DatabaseDSN   string
	Key           string
	AuditFile     string
	AuditURL      string
	CryptoKey     string
	ConfigFile    string
}

func parseServerFlags() (*flag.FlagSet, serverFlags, error) {
	var f serverFlags

	fs := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)

	fs.StringVar(&f.RunAddr, "a", "", "address and port to run server")
	fs.StringVar(&f.FilePath, "f", "", "file path to write metrics")
	fs.BoolVar(&f.Restore, "r", false, "load metrics from storage")

	fs.DurationVar(&f.StoreInterval, "i", 0, "store interval in seconds")

	fs.StringVar(&f.DatabaseDSN, "d", "", "address to connect DataBase")

	fs.StringVar(&f.Key, "k", "", "signing key")

	fs.StringVar(&f.AuditFile, "audit-file", "", "file path to write audit")
	fs.StringVar(&f.AuditURL, "audit-url", "", "addres to send audit")

	fs.StringVar(&f.CryptoKey, "crypto-key", "", "file path to private key for encryption")

	fs.StringVar(&f.ConfigFile, "c", "", "path to config file")
	fs.StringVar(&f.ConfigFile, "config", "", "path to config file")

	if err := fs.Parse(os.Args[1:]); err != nil {
		return nil, f, err
	}
	return fs, f, nil
}

type serverConfigJSON struct {
	RunAddr       *string `json:"address"`
	StoreInterval *string `json:"store_interval"`
	FilePath      *string `json:"store_file"`
	Restore       *bool   `json:"restore"`
	DatabaseDSN   *string `json:"database_dsn"`
	CryptoKey     *string `json:"crypto_key"`
}

func applyJSON(cfg *ServerConfig, jc serverConfigJSON) error {
	if jc.RunAddr != nil {
		cfg.RunAddr = *jc.RunAddr
	}

	if jc.StoreInterval != nil {

		d, err := time.ParseDuration(*jc.StoreInterval)
		if err != nil {
			return fmt.Errorf("invalid store_interval: %w", err)
		}
		cfg.StoreInterval = d
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

	return nil
}

func applyFlags(cfg *ServerConfig, fs *flag.FlagSet, f serverFlags) {
	fs.Visit(func(fl *flag.Flag) {
		switch fl.Name {
		case "a":
			cfg.RunAddr = f.RunAddr
		case "i":
			cfg.StoreInterval = f.StoreInterval
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
