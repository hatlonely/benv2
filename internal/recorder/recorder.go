package recorder

import (
	"reflect"
	"time"

	"github.com/hatlonely/go-kit/refx"
	"github.com/pkg/errors"
)

func RegisterRecorder(key string, constructor interface{}) {
	refx.Register("recorder.Recorder", key, constructor)
}

func NewRecorderWithOptions(options *refx.TypeOptions, opts ...refx.Option) (Recorder, error) {
	if options.Namespace == "" {
		options.Namespace = "recorder.Recorder"
	}
	v, err := refx.NewType(reflect.TypeOf((*Recorder)(nil)).Elem(), options, opts...)
	if err != nil {
		return nil, errors.WithMessage(err, "refx.NewType failed")
	}

	return v.(Recorder), nil
}

type Recorder interface {
	RecordMeta(meta *Meta) error
	Record(stat *UnitStat) error
	Close() error
}

type Meta struct {
	ID        string
	Name      string
	Parallel  []map[string]int
	TimeRange []*struct {
		StartTime time.Time
		EndTime   time.Time
	}
}

type UnitStat struct {
	ID      string
	Time    string
	Seq     int
	Name    string
	Step    []*StepStat
	ErrCode string
	ResTime time.Duration
}

type StepStat struct {
	Time    string
	Req     interface{}
	Res     interface{}
	Err     error
	ErrCode string
	ResTime time.Duration
}
