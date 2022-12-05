package recorder

import (
	"reflect"
	"time"

	"github.com/hatlonely/go-kit/refx"
	"github.com/pkg/errors"
)

func RegisterAnalyst(key string, constructor interface{}) {
	refx.Register("recorder.Analyst", key, constructor)
}

func NewAnalystWithOptions(options *refx.TypeOptions, opts ...refx.Option) (Analyst, error) {
	if options.Namespace == "" {
		options.Namespace = "recorder.Analyst"
	}
	v, err := refx.NewType(reflect.TypeOf((*Analyst)(nil)).Elem(), options, opts...)
	if err != nil {
		return nil, errors.WithMessage(err, "refx.NewType failed")
	}

	return v.(Analyst), nil
}

type Analyst interface {
	TimeRange() (time.Time, time.Time, error)
	UnitStatStream() StatStream
}

type StatStream interface {
	Next() (*UnitStat, error)
}
