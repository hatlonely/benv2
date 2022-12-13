package reporter

import (
	"github.com/hatlonely/benv2/internal/recorder"
	"github.com/hatlonely/go-kit/strx"
)

type TextReporter struct{}

func (r *TextReporter) Report(metric []*recorder.Metric) string {
	return strx.JsonMarshalIndentSortKeys(metric)
}
