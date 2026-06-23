package buildinfo

import "log"

// Глоальные переменные заполняются во время сборки с помощью параметра -ldflags.
// Пример: go run -ldflags "-X 'github.com/onbehalfofhim/metric-alert/internal/buildinfo.Version=1.0.0'" ./cmd/server/main.go
// Значения по умолчанию пусты; при выводе на экран они будут отображаться как "N/A".
var (
	Version string
	Date    string
	Commit  string
)

// Вывод информации о сборке в стандартный поток вывода (через пакет log).
func Print() {
	version := Version
	if version == "" {
		version = "N/A"
	}

	date := Date
	if date == "" {
		date = "N/A"
	}

	commit := Commit
	if commit == "" {
		commit = "N/A"
	}

	log.Printf("Build version: %s", version)
	log.Printf("Build date: %s", date)
	log.Printf("Build commit: %s", commit)
}
