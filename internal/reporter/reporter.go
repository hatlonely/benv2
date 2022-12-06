package reporter

import (
	"reflect"

	"github.com/hatlonely/go-kit/refx"
	"github.com/pkg/errors"
)

func RegisterReporter(key string, constructor interface{}) {
	refx.Register("reporter.Reporter", key, constructor)
}

func NewReporterWithOptions(options *refx.TypeOptions, opts ...refx.Option) (Reporter, error) {
	if options.Namespace == "" {
		options.Namespace = "reporter.Reporter"
	}
	v, err := refx.NewType(reflect.TypeOf((*Reporter)(nil)).Elem(), options, opts...)
	if err != nil {
		return nil, errors.WithMessage(err, "refx.NewType failed")
	}

	return v.(Reporter), nil
}

type Reporter interface {
	Report() string
}
