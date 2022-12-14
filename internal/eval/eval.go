package eval

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"
	"unsafe"

	"github.com/PaesslerAG/gval"
	"github.com/barkimedes/go-deepcopy"
	"github.com/hatlonely/go-kit/refx"
	jsoniter "github.com/json-iterator/go"
	"github.com/pkg/errors"
)

func init() {
	// 数值转换成 int64，默认都是 float64
	jsoniter.RegisterTypeDecoderFunc("interface {}", func(ptr unsafe.Pointer, iter *jsoniter.Iterator) {
		switch iter.WhatIsNext() {
		case jsoniter.NumberValue:
			var number json.Number
			iter.ReadVal(&number)
			i, err := strconv.ParseInt(string(number), 10, 64)
			if err == nil {
				*(*interface{})(ptr) = i
				return
			}
			f, err := strconv.ParseFloat(string(number), 64)
			if err == nil {
				*(*interface{})(ptr) = f
				return
			}
		default:
			*(*interface{})(ptr) = iter.Read()
		}
	})
}

func NewEvaluable(v interface{}) (*Evaluable, error) {
	ev := &Evaluable{
		consts:    nil,
		variables: map[string]gval.Evaluable{},
	}

	err := refx.InterfaceTravel(v, func(key string, val interface{}) error {
		idx := strings.LastIndexByte(key, '.')
		if idx+2 < len(key) && key[idx+1] == '#' {
			expr, ok := val.(string)
			if !ok {
				return errors.Errorf("expression should be string. key: [%s]", key)
			}
			e, err := Lang.NewEvaluable(expr)
			if err != nil {
				return errors.Wrap(err, "Lang.NewEvaluable failed")
			}

			ev.variables[key[0:idx+1]+key[idx+2:]] = e
		} else {
			if err := refx.InterfaceSet(&ev.consts, key, val); err != nil {
				return errors.WithMessage(err, "refx.InterfaceSet failed")
			}
		}

		return nil
	})

	if err != nil {
		return nil, errors.Wrap(err, "refx.InterfaceTravel failed")
	}

	return ev, nil
}

type Evaluable struct {
	consts    interface{}
	variables map[string]gval.Evaluable
}

func (e *Evaluable) Evaluate(vals interface{}) (interface{}, error) {
	v := deepCopy(e.consts)
	for key, evaluable := range e.variables {
		val, err := evaluable(context.Background(), vals)
		if err != nil {
			return nil, errors.Wrap(err, "evaluable failed")
		}
		if err := refx.InterfaceSet(&v, key, val); err != nil {
			return nil, errors.Wrap(err, "refx.InterfaceSet failed")
		}
	}

	return v, nil
}

func deepCopy(src interface{}) interface{} {
	return deepCopyByGoDeepcopy(src)
}

func deepCopyByRefxSet(src interface{}) interface{} {
	var dst interface{}

	_ = refx.InterfaceTravel(src, func(key string, val interface{}) error {
		_ = refx.InterfaceSet(&dst, key, val)
		return nil
	})

	return dst
}

func deepCopyByJsonMarshal(src interface{}) interface{} {
	buf, _ := jsoniter.Marshal(src)

	var v interface{}
	_ = jsoniter.Unmarshal(buf, &v)

	return v
}

func deepCopyByGoDeepcopy(src interface{}) interface{} {
	dst, _ := deepcopy.Anything(src)
	return dst
}
