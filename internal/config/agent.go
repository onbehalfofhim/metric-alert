package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/caarlos0/env/v11"
)

type AgentConfig struct {
	RunAddr        string        `env:"ADDRESS"`
	PollInterval   time.Duration `env:"POLL_INTERVAL"`
	ReportInterval time.Duration `env:"REPORT_INTERVAL"`
	Key            string        `env:"KEY"`
	RateLimit      int           `env:"RATE_LIMIT"`
	CryptoKey      string        `env:"CRYPTO_KEY"`
	ConfigFile     string        `env:"CONFIG"`
	GRPCAddr       string        `env:"GRPC_ADDRESS"`
}

func defaultAgentConfig() AgentConfig {
	return AgentConfig{
		RunAddr:        "localhost:8080",
		PollInterval:   2 * time.Second,
		ReportInterval: 10 * time.Second,
		RateLimit:      1,
		GRPCAddr:       "localhost:3200",
	}
}

// обработка аргументов командной строки
// и сохранение их значения в структуре
func ParseAgentFlags() (AgentConfig, error) {
	// проставялем дефолтные значения
	cfg := defaultAgentConfig()

	/// парсим переданные серверу аргументы командной строки в зарегистрированные переменные
	fs, flags, err := parseAgentFlags()
	if err != nil {
		return cfg, err
	}

	// CONFIG можно задать через env
	if flags.ConfigFile == "" {
		flags.ConfigFile = os.Getenv("CONFIG")
	}

	if flags.ConfigFile != "" {
		var jc agentConfigJSON
		data, err := os.ReadFile(flags.ConfigFile)
		if err != nil {
			return cfg, err
		}
		if err := json.Unmarshal(data, &jc); err != nil {
			return cfg, err
		}

		if err := applyAgentJSON(&cfg, jc); err != nil {
			return cfg, err

		}
	}

	// применяем явно переданные флаги поверх JSON-файла
	applyAgentFlags(&cfg, fs, flags)

	// парсим переменные окружения

	if err := env.Parse(&cfg); err != nil {
		return cfg, fmt.Errorf("can't parse environment variables: %w", err)
	}

	if !strings.HasPrefix(cfg.RunAddr, "http://") {
		cfg.RunAddr = "http://" + cfg.RunAddr
	}

	return cfg, nil
}

type agentFlags struct {
	RunAddr        string
	PollInterval   time.Duration
	ReportInterval time.Duration
	Key            string
	RateLimit      int
	CryptoKey      string
	ConfigFile     string
	GRPCAddr       string
}

func parseAgentFlags() (*flag.FlagSet, agentFlags, error) {
	var f agentFlags

	fs := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)

	fs.StringVar(&f.RunAddr, "a", "", "address and port to run server")
	fs.DurationVar(&f.PollInterval, "p", 0, "frequency of polling metrics from the runtime package")
	fs.DurationVar(&f.ReportInterval, "r", 0, "frequency of sending metrics to the server")
	fs.StringVar(&f.Key, "k", "", "signing key")
	fs.IntVar(&f.RateLimit, "l", 0, "the number of simultaneously outgoing requests to the server")

	fs.StringVar(&f.CryptoKey, "crypto-key", "", "file path to public key for encryption")

	fs.StringVar(&f.ConfigFile, "c", "", "path to config file")
	fs.StringVar(&f.ConfigFile, "config", "", "path to config file")

	fs.StringVar(&f.GRPCAddr, "g", "", "gRPC server address")

	if err := fs.Parse(os.Args[1:]); err != nil {
		return nil, f, err
	}
	return fs, f, nil
}

type agentConfigJSON struct {
	RunAddr        *string `json:"address"`
	PollInterval   *string `json:"poll_interval"`
	ReportInterval *string `json:"report_interval"`
	CryptoKey      *string `json:"crypto_key"`
}

func applyAgentJSON(cfg *AgentConfig, jc agentConfigJSON) error {
	if jc.RunAddr != nil {
		cfg.RunAddr = *jc.RunAddr
	}

	if jc.PollInterval != nil {
		d, err := time.ParseDuration(*jc.PollInterval)
		if err != nil {
			return fmt.Errorf("invalid poll_interval: %w", err)
		}
		cfg.PollInterval = d
	}

	if jc.ReportInterval != nil {
		d, err := time.ParseDuration(*jc.ReportInterval)
		if err != nil {
			return fmt.Errorf("invalid report_interval: %w", err)
		}
		cfg.ReportInterval = d
	}

	if jc.CryptoKey != nil {
		cfg.CryptoKey = *jc.CryptoKey
	}

	return nil
}

func applyAgentFlags(cfg *AgentConfig, fs *flag.FlagSet, f agentFlags) {
	fs.Visit(func(fl *flag.Flag) {
		switch fl.Name {
		case "a":
			cfg.RunAddr = f.RunAddr
		case "p":
			cfg.PollInterval = f.PollInterval
		case "r":
			cfg.ReportInterval = f.ReportInterval
		case "crypto-key":
			cfg.CryptoKey = f.CryptoKey
		case "g":
			cfg.GRPCAddr = f.GRPCAddr
		}
	})
}
