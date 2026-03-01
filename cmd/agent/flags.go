package main

import (
	"flag"
)

// неэкспортированная переменная flagRunAddr содержит адрес и порт для запуска сервера
var flagRunAddr string
var pollInterval int64
var reportInterval int64

// parseFlags обрабатывает аргументы командной строки
// и сохраняет их значения в соответствующих переменных
func parseFlags() {
	// регистрируем переменные
	// как аргументы со значениями по умолчанию
	flag.StringVar(&flagRunAddr, "a", "localhost:8080", "address and port to run server")
	flag.Int64Var(&pollInterval, "p", 2, "frequency of polling metrics from the runtime package")
	flag.Int64Var(&reportInterval, "r", 10, "frequency of sending metrics to the server")

	// парсим переданные серверу аргументы в зарегистрированные переменные
	flag.Parse()
}
