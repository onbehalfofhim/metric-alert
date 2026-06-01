package config

import (
	"flag"
	"fmt"
	"strings"

	"github.com/caarlos0/env/v11"
)

type AgentConfig struct {
	RunAddr        string `env:"ADDRESS"`
	PollInterval   int    `env:"REPORT_INTERVAL"`
	ReportInterval int    `env:"POLL_INTERVAL"`
	Key            string `env:"KEY"`
	RateLimit      int    `env:"RATE_LIMIT"`
}

// обработка аргументов командной строки
// и сохранение их значения в структуре
func ParseAgentFlags() (AgentConfig, error) {
	var cfg AgentConfig

	// регистрируем переменные
	// как аргументы со значениями по умолчанию
	flag.StringVar(&cfg.RunAddr, "a", "localhost:8080", "address and port to run server")
	flag.IntVar(&cfg.PollInterval, "p", 2, "frequency of polling metrics from the runtime package")
	flag.IntVar(&cfg.ReportInterval, "r", 10, "frequency of sending metrics to the server")
	flag.StringVar(&cfg.Key, "k", "", "signing key")
	flag.IntVar(&cfg.RateLimit, "l", 1, "the number of simultaneously outgoing requests to the server")

	// парсим переданные серверу аргументы командной строки в зарегистрированные переменные
	flag.Parse()

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
