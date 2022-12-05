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
	Record(stat *UnitStat) error
	Close() error

	TimeRange() (time.Time, time.Time, error)
	//Stat() StatStream
}

type StatStream interface {
	Next() *UnitStat
}

type UnitStat struct {
	Time    string // 记录时间，由 Recorder 设置
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
