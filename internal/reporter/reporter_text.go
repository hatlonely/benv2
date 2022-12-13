package reporter

import (
	"bytes"
	"fmt"
	"sort"
	"strings"

	"github.com/hatlonely/benv2/internal/recorder"
)

type TextReporter struct{}

func (r *TextReporter) Report(meta *recorder.Meta, metrics []*recorder.Metric) string {
	var buf bytes.Buffer

	for i := range metrics {
		buf.WriteString(buildUnit(meta.Parallel[i], metrics[i]))
		buf.WriteString("----------------------------------\n")
	}

	return buf.String()
}

func buildUnit(parallel map[string]int, metric *recorder.Metric) string {
	var buf bytes.Buffer

	buf.WriteString(buildParallel(parallel))
	buf.WriteString("\n")

	buf.WriteString(buildMeasurementMap("QPS", metric.QPS))

	return buf.String()
}

func buildParallel(parallel map[string]int) string {
	var keys []string
	for key := range parallel {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	var parallels []string
	for _, key := range keys {
		parallels = append(parallels, fmt.Sprintf("%s: %d", key, parallel[key]))
	}

	return strings.Join(parallels, ",")
}

func buildMeasurementMap(title string, measurementMap map[string][]*recorder.Measurement) string {
	var buf bytes.Buffer

	keyWidth := len(title)
	var keys []string
	for key := range measurementMap {
		keys = append(keys, key)
		if len(key) > keyWidth {
			keyWidth = len(key)
		}
	}
	sort.Strings(keys)

	// |QPS  |10:01|10:02|10:02|10:03|10:03|10:03|
	buf.WriteByte('|')
	buf.WriteString(title)
	for i := len(title); i < keyWidth; i++ {
		buf.WriteByte(' ')
	}
	buf.WriteByte('|')
	for _, measurements := range measurementMap[keys[0]] {
		buf.WriteString(measurements.Time.Format("04:05"))
		buf.WriteByte('|')
	}
	buf.WriteByte('\n')

	// |-----|-----|-----|-----|-----|-----|-----|
	buf.WriteByte('|')
	for i := 0; i < keyWidth; i++ {
		buf.WriteByte('-')
	}
	buf.WriteByte('|')
	for range measurementMap[keys[0]] {
		buf.WriteString("-----")
		buf.WriteByte('|')
	}
	buf.WriteByte('\n')

	// |QPS  |10:01|10:02|10:02|10:03|10:03|10:03|
	for _, key := range keys {
		buf.WriteByte('|')
		buf.WriteString(key)
		for i := len(key); i < keyWidth; i++ {
			buf.WriteByte(' ')
		}
		buf.WriteByte('|')
		for _, measurement := range measurementMap[key] {
			buf.WriteString(fmt.Sprintf("%.1f", measurement.Value))
			buf.WriteByte('|')
		}
		buf.WriteByte('\n')
	}

	return buf.String()
}
