package templates

import (
	_ "embed"
	"html/template"
	"io"
)

//go:embed metrics.html
var metricsHTML string

var metricsTmpl = template.Must(
	template.New("metrics").Parse(metricsHTML),
)

type MetricView struct {
	Name  string
	Value string
}

type MetricsPageData struct {
	Gauges   []MetricView
	Counters []MetricView
}

func RenderMetricsPage(w io.Writer, data MetricsPageData) error {
	return metricsTmpl.Execute(w, data)
}
