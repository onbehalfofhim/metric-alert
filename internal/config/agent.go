package config

import (
	"flag"
	"strings"
)

type AgentConfig struct {
	RunAddr        string
	PollInterval   int
	ReportInterval int
}

// обработка аргументов командной строки
// и сохранение их значения в структуре
func ParseAgentFlags() AgentConfig {
	var cfg AgentConfig

	// регистрируем переменные
	// как аргументы со значениями по умолчанию
	flag.StringVar(&cfg.RunAddr, "a", "localhost:8080", "address and port to run server")
	flag.IntVar(&cfg.PollInterval, "p", 2, "frequency of polling metrics from the runtime package")
	flag.IntVar(&cfg.ReportInterval, "r", 10, "frequency of sending metrics to the server")

	// парсим переданные серверу аргументы в зарегистрированные переменные
	flag.Parse()

	if !strings.HasPrefix(cfg.RunAddr, "http://") {
		cfg.RunAddr = "http://" + cfg.RunAddr
	}

	return cfg
}
