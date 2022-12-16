package reporter

import (
	"github.com/hatlonely/benv2/internal/recorder"
)

type HtmlReporterOptions struct {
}

func NewHtmlReporterWithOptions(options *HtmlReporterOptions) *HtmlReporter {
	return &HtmlReporter{
		options: options,
	}
}

type HtmlReporter struct {
	options *HtmlReporterOptions
}

func (r *HtmlReporter) Report(meta *recorder.Meta, metrics []*recorder.Metric) string {
	return "html"
}
