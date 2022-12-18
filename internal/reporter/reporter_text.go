package reporter

import (
	"bytes"
	"fmt"
	"sort"
	"strings"

	"github.com/hatlonely/go-kit/strx"
	"github.com/spf13/cast"

	"github.com/hatlonely/benv2/internal/recorder"
)

type TextReporterOptions struct {
	TitleWidth int `dft:"20"`
	ValueWidth int `dft:"9"`
}

func NewTextReporterWithOptions(options *TextReporterOptions) *TextReporter {
	return &TextReporter{
		options: options,
	}
}

type TextReporter struct {
	options *TextReporterOptions
}

func (r *TextReporter) Report(meta *recorder.Meta, metrics []*recorder.Metric, monitors []map[string][]*recorder.Measurement) string {
	var buf bytes.Buffer

	for i := range metrics {
		buf.WriteString(r.buildUnit(meta.Parallel[i], metrics[i], monitors[i]))
		buf.WriteString("==================================================================================\n")
	}

	return buf.String()
}

func (r *TextReporter) buildUnit(parallel map[string]int, metric *recorder.Metric, monitor_ map[string][]*recorder.Measurement) string {
	var buf bytes.Buffer

	buf.WriteString(buildParallel(parallel))
	buf.WriteByte('\n')

	buf.WriteString(buildSummary(r.options.TitleWidth, metric.Summary))
	buf.WriteByte('\n')

	buf.WriteString(buildErrCodeDistribution(metric.ErrCodeDistribution))
	buf.WriteByte('\n')
	buf.WriteString(buildMeasurementMap(r.options.TitleWidth, r.options.ValueWidth, "QPS", metric.QPS))
	buf.WriteByte('\n')
	buf.WriteString(buildMeasurementMap(r.options.TitleWidth, r.options.ValueWidth, "AvgResTimeMs", metric.AvgResTimeMs))
	buf.WriteByte('\n')
	buf.WriteString(buildMeasurementMap(r.options.TitleWidth, r.options.ValueWidth, "SuccessRatePercent", metric.SuccessRatePercent))
	buf.WriteByte('\n')

	for key, val := range monitor_ {
		buf.WriteString(buildMeasurementMap(r.options.TitleWidth, r.options.ValueWidth, key, map[string][]*recorder.Measurement{key: val}))
		buf.WriteByte('\n')
	}

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

	return strings.Join(parallels, ", ") + "\n"
}

func buildSummary(titleWidth int, summaryMap map[string]*recorder.Summary) string {
	var buf bytes.Buffer

	width := titleWidth
	if width < len("summary") {
		width = len("summary")
	}
	var keys []string
	for key := range summaryMap {
		keys = append(keys, key)
		if len(key) > width {
			width = len(key)
		}
	}
	sort.Strings(keys)

	titles := []string{
		"    Total    ",
		"     QPS     ",
		"AvgResTimeMs",
		"SuccessRatePercent",
	}
	buf.WriteByte('|')
	appendCenter(&buf, width, "summary")
	buf.WriteByte('|')
	for _, title := range titles {
		appendCenter(&buf, len(title)+2, title)
		buf.WriteByte('|')
	}
	buf.WriteByte('\n')
	buf.WriteByte('|')
	for i := 0; i < width; i++ {
		buf.WriteByte('-')
	}
	buf.WriteByte('|')
	for _, title := range titles {
		for i := 0; i < len(title)+2; i++ {
			buf.WriteByte('-')
		}
		buf.WriteByte('|')
	}
	buf.WriteByte('\n')

	for _, key := range keys {
		buf.WriteByte('|')
		appendCenter(&buf, width, key)
		buf.WriteByte('|')
		appendCenter(&buf, len(titles[0])+2, cast.ToString(summaryMap[key].Total))
		buf.WriteByte('|')
		appendCenter(&buf, len(titles[1])+2, fmt.Sprintf("%.2f", summaryMap[key].QPS))
		buf.WriteByte('|')
		appendCenter(&buf, len(titles[2])+2, fmt.Sprintf("%.2f", summaryMap[key].AvgResTimeMs))
		buf.WriteByte('|')
		appendCenter(&buf, len(titles[3])+2, fmt.Sprintf("%.2f", summaryMap[key].SuccessRatePercent))
		buf.WriteByte('|')
		buf.WriteByte('\n')
	}

	return buf.String()
}

func buildErrCodeDistribution(errCodeDistribution map[string]map[string]int) string {
	var buf bytes.Buffer

	for key, val := range errCodeDistribution {
		buf.WriteString(key)
		buf.WriteString(": ")
		buf.WriteString(strx.JsonMarshalSortKeys(val))
		buf.WriteByte('\n')
	}

	return buf.String()
}

func appendCenter(buf *bytes.Buffer, width int, str string) {
	if len(str) >= width {
		buf.WriteString(str)
		return
	}
	space := width - len(str)
	left := space / 2
	right := space - left
	for i := 0; i < left; i++ {
		buf.WriteByte(' ')
	}
	buf.WriteString(str)
	for i := 0; i < right; i++ {
		buf.WriteByte(' ')
	}
}

func buildMeasurementMap(titleWidth, valueWidth int, title string, measurementMap map[string][]*recorder.Measurement) string {
	var buf bytes.Buffer

	keyWidth := titleWidth
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
	appendCenter(&buf, keyWidth, title)
	buf.WriteByte('|')
	for _, measurements := range measurementMap[keys[0]] {
		appendCenter(&buf, valueWidth, measurements.Time.Format("04:05"))
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
		for i := 0; i < valueWidth; i++ {
			buf.WriteByte('-')
		}
		buf.WriteByte('|')
	}
	buf.WriteByte('\n')

	// |QPS  |10:01|10:02|10:02|10:03|10:03|10:03|
	for _, key := range keys {
		buf.WriteByte('|')
		appendCenter(&buf, keyWidth, key)
		buf.WriteByte('|')
		for _, measurement := range measurementMap[key] {
			appendCenter(&buf, valueWidth, fmt.Sprintf("%.2f", measurement.Value))
			buf.WriteByte('|')
		}
		buf.WriteByte('\n')
	}

	return buf.String()
}
