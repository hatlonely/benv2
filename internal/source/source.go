package source

import (
	"reflect"

	"github.com/hatlonely/go-kit/refx"
	"github.com/pkg/errors"
)

func RegisterSource(key string, constructor interface{}) {
	refx.Register("source.Source", key, constructor)
}

func NewSourceWithOptions(options *refx.TypeOptions, opts ...refx.Option) (Source, error) {
	if options.Namespace == "" {
		options.Namespace = "source.Source"
	}
	v, err := refx.NewType(reflect.TypeOf((*Source)(nil)).Elem(), options, opts...)
	if err != nil {
		return nil, errors.WithMessage(err, "refx.NewType failed")
	}

	return v.(Source), nil
}

type Source interface {
	Fetch() interface{}
}
