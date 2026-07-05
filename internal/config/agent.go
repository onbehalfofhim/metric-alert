package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/caarlos0/env/v11"
)

type AgentConfig struct {
	RunAddr        string `env:"ADDRESS"`
	PollInterval   int    `env:"POLL_INTERVAL"`
	ReportInterval int    `env:"REPORT_INTERVAL"`
	Key            string `env:"KEY"`
	RateLimit      int    `env:"RATE_LIMIT"`
	CryptoKey      string `env:"CRYPTO_KEY"`
	ConfigFile     string `env:"CONFIG"`
}

func defaultAgentConfig() AgentConfig {
	return AgentConfig{
		RunAddr:        "localhost:8080",
		PollInterval:   2,
		ReportInterval: 10,
		RateLimit:      1,
	}
}

// обработка аргументов командной строки
// и сохранение их значения в структуре
func ParseAgentFlags() (AgentConfig, error) {
	// проставялем дефолтные значения
	cfg := defaultAgentConfig()

	/// парсим переданные серверу аргументы командной строки в зарегистрированные переменные
	flags := parseAgentFlags()

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
		applyAgentJSON(&cfg, jc)
	}

	// применяем явно переданные флаги поверх JSON-файла
	applyAgentFlags(&cfg, flags)

	// парсим переменные окружения
	err := env.Parse(&cfg)
	if err != nil {
		return cfg, fmt.Errorf("can't parse environment variables: %w", err)
	}

	if !strings.HasPrefix(cfg.RunAddr, "http://") {
		cfg.RunAddr = "http://" + cfg.RunAddr
	}

	return cfg, nil
}

type agentFlags struct {
	RunAddr        string
	PollInterval   int
	ReportInterval int
	Key            string
	RateLimit      int
	CryptoKey      string
	ConfigFile     string
}

func parseAgentFlags() agentFlags {
	var f agentFlags

	flag.StringVar(&f.RunAddr, "a", "", "address and port to run server")
	flag.IntVar(&f.PollInterval, "p", 0, "frequency of polling metrics from the runtime package")
	flag.IntVar(&f.ReportInterval, "r", 0, "frequency of sending metrics to the server")
	flag.StringVar(&f.Key, "k", "", "signing key")
	flag.IntVar(&f.RateLimit, "l", 0, "the number of simultaneously outgoing requests to the server")

	flag.StringVar(&f.CryptoKey, "crypto-key", "", "file path to public key for encryption")

	flag.StringVar(&f.ConfigFile, "c", "", "path to config file")
	flag.StringVar(&f.ConfigFile, "config", "", "path to config file")

	flag.Parse()

	return f
}

type agentConfigJSON struct {
	RunAddr        *string `json:"address"`
	PollInterval   *int    `json:"poll_interval"`
	ReportInterval *int    `json:"report_interval"`
	CryptoKey      *string `json:"crypto_key"`
}

func applyAgentJSON(cfg *AgentConfig, jc agentConfigJSON) {
	if jc.RunAddr != nil {
		cfg.RunAddr = *jc.RunAddr
	}
	if jc.PollInterval != nil {
		cfg.PollInterval = *jc.PollInterval
	}
	if jc.ReportInterval != nil {
		cfg.ReportInterval = *jc.ReportInterval
	}
	if jc.CryptoKey != nil {
		cfg.CryptoKey = *jc.CryptoKey
	}
}

func applyAgentFlags(cfg *AgentConfig, f agentFlags) {
	flag.Visit(func(fl *flag.Flag) {
		switch fl.Name {
		case "a":
			cfg.RunAddr = f.RunAddr

		case "p":
			cfg.PollInterval = f.PollInterval

		case "r":
			cfg.ReportInterval = f.ReportInterval

		case "crypto-key":
			cfg.CryptoKey = f.CryptoKey
		}
	})
}
