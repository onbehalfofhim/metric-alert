package config

import (
	"flag"
)

type ServerConfig struct {
	RunAddr string
}

// обработка аргументов командной строки
// и сохраняет их значения в структуре
func ParseServerFlags() ServerConfig {
	var cfg ServerConfig

	// регистрируем переменную RunAddr
	// как аргумент -a со значением по умолчанию
	flag.StringVar(&cfg.RunAddr, "a", "localhost:8080", "address and port to run server")
	// парсим переданные серверу аргументы в зарегистрированные переменные
	flag.Parse()

	return cfg
}
