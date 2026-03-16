package config

import (
	"flag"
	"log"

	"github.com/caarlos0/env/v6"
)

type ServerConfig struct {
	RunAddr string `env:"ADDRESS"`
}

// обработка аргументов командной строки
// и сохраняет их значения в структуре
func ParseServerFlags() ServerConfig {
	var cfg ServerConfig

	// регистрируем переменную RunAddr
	// как аргумент -a со значением по умолчанию
	flag.StringVar(&cfg.RunAddr, "a", "localhost:8080", "address and port to run server")
	// парсим переданные серверу аргументы командной строки в зарегистрированные переменные
	flag.Parse()

	// парсим переменные окружения
	err := env.Parse(&cfg)
	if err != nil {
		log.Fatal(err)
	}

	return cfg
}
