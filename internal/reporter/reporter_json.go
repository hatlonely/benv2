package reporter

import (
	"github.com/hatlonely/go-kit/strx"

	"github.com/hatlonely/benv2/internal/recorder"
)

type JsonReporter struct{}

func (r *JsonReporter) Report(meta *recorder.Meta, metrics []*recorder.Metric) string {
	return strx.JsonMarshalIndentSortKeys(map[string]interface{}{
		"Meta":    meta,
		"Metrics": metrics,
	})
}
