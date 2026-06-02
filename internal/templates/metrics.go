package templates

import (
	"io"
	"strings"
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
	var b strings.Builder

	b.Grow(len(data.Gauges)*50 + len(data.Counters)*50)

	b.WriteString(`<!DOCTYPE html>
		<html>
		<head>
		<title>Metrics</title>
		</head>
		<body>
		<h1>Current Metrics</h1>
		<table border="1">
		<thead>
		<tr>
		<th>Name</th>
		<th>Value</th>
		</tr>
		</thead>
		<tbody>
	`)

	b.WriteString(`<tr><td colspan="2">gauges</td></tr>`)

	for _, m := range data.Gauges {
		b.WriteString("<tr><td>")
		b.WriteString(m.Name)
		b.WriteString("</td><td>")
		b.WriteString(m.Value)
		b.WriteString("</td></tr>")
	}

	b.WriteString(`<tr><td colspan="2">counters</td></tr>`)

	for _, m := range data.Counters {
		b.WriteString("<tr><td>")
		b.WriteString(m.Name)
		b.WriteString("</td><td>")
		b.WriteString(m.Value)
		b.WriteString("</td></tr>")
	}

	b.WriteString(`
		</tbody>
		</table>
		</body>
		</html>
	`)

	_, err := io.WriteString(w, b.String())
	return err
}
