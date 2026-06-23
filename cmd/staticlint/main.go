/*
Command staticlint запускает набор статических анализаторов.
Состав анализаторов:
  - стандартные анализаторы golang.org/x/tools/go/analysis/passes:
    –- printf: проверяет корректность форматных строк и аргументов функций семейства fmt.Printf;
    –- shadow: обнаруживает затенение переменных во вложенных областях видимости;
    –- structtag: проверяет корректность тегов структур и их соответствие соглашениям reflect;
    –- atomic: выявляет типичные ошибки при использовании пакета sync/atomic;
  - анализаторы класса SA из staticcheck:
    –- выполняют поиск потенциальных ошибок, которые могут приводить к
    некорректной работе программы, паникам, утечкам ресурсов и другим дефектам;
  - анализаторы класса S из staticcheck/simple:
    –- выявляют участки кода, которые можно упростить без изменения поведения программы;
  - errcheck:
    -- проверяет, что не игнорируются возвращаемые ошибки;
  - ineffassign:
    –- обнаруживает присваивания переменным, результаты которых никогда не используются;
  - exitcheck:
    -- запрещает прямой вызов os.Exit в функции main пакета main;

Запуск:
go run ./cmd/staticlint/main.go ./...
*/
package main

import (
	"strings"

	"github.com/gordonklaus/ineffassign/pkg/ineffassign"
	"github.com/kisielk/errcheck/errcheck"
	"github.com/onbehalfofhim/metric-alert/cmd/staticlint/exitcheck"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/atomic"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"honnef.co/go/tools/simple"
	"honnef.co/go/tools/staticcheck"
)

func main() {
	var analyzers = []*analysis.Analyzer{
		printf.Analyzer,
		shadow.Analyzer,
		structtag.Analyzer,
		atomic.Analyzer,
	}

	// SA analyzers
	for _, v := range staticcheck.Analyzers {
		if strings.HasPrefix(v.Analyzer.Name, "SA") {
			analyzers = append(analyzers, v.Analyzer)
		}

	}

	// non-SA analyzers
	for _, a := range simple.Analyzers {
		analyzers = append(analyzers, a.Analyzer)
	}

	// public and exit-check analyzers
	analyzers = append(analyzers,
		errcheck.Analyzer,
		ineffassign.Analyzer,
		exitcheck.Analyzer,
	)

	multichecker.Main(analyzers...)
}
