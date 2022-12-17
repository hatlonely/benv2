package monitor

import (
	"reflect"
	"time"

	"github.com/hatlonely/go-kit/refx"
	"github.com/pkg/errors"

	"github.com/hatlonely/benv2/internal/recorder"
)

func RegisterMonitor(key string, constructor interface{}) {
	refx.Register("monitor.Monitor", key, constructor)
}

func NewMonitorWithOptions(options *refx.TypeOptions, opts ...refx.Option) (Monitor, error) {
	if options.Namespace == "" {
		options.Namespace = "monitor.Monitor"
	}
	v, err := refx.NewType(reflect.TypeOf((*Monitor)(nil)).Elem(), options, opts...)
	if err != nil {
		return nil, errors.WithMessage(err, "refx.NewType failed")
	}
	return v.(Monitor), nil
}

type Monitor interface {
	Collect(startTime time.Time, endTime time.Time) (map[string][]*recorder.Measurement, error)
}
