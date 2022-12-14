package driver

import (
	"context"
	"encoding/json"
	"reflect"
	"strconv"
	"unsafe"

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

func RegisterDriver(key string, constructor interface{}) {
	refx.Register("driver.Driver", key, constructor)
}

func NewDriverWithOptions(options *refx.TypeOptions, opts ...refx.Option) (Driver, error) {
	if options.Namespace == "" {
		options.Namespace = "driver.Driver"
	}
	v, err := refx.NewType(reflect.TypeOf((*Driver)(nil)).Elem(), options, opts...)
	if err != nil {
		return nil, errors.WithMessage(err, "refx.NewType failed")
	}

	return v.(Driver), nil
}

type Driver interface {
	Do(req interface{}) (interface{}, error)
}

type WrapDriver struct {
	//inner     interface{}
	inner      reflect.Value
	methodKey  string
	methodName string
}

func NewWrapDriverWithMethodKey(v interface{}, methodKey string) func(options interface{}) (*WrapDriver, error) {
	constructor, err := refx.NewConstructor(v)
	refx.Must(err)
	return func(options interface{}) (*WrapDriver, error) {
		result, err := constructor.Call(options)
		if err != nil {
			return nil, errors.WithMessage(err, "constructor.Call failed")
		}

		if constructor.ReturnError {
			if !result[1].IsNil() {
				return nil, errors.Wrapf(result[1].Interface().(error), "New failed")
			}
			return &WrapDriver{
				inner:     result[0],
				methodKey: methodKey,
			}, nil
		}

		return &WrapDriver{
			inner:     result[0],
			methodKey: methodKey,
		}, nil
	}
}

func NewWrapDriverWithMethodName(v interface{}, methodName string) func(options interface{}) (*WrapDriver, error) {
	constructor, err := refx.NewConstructor(v)
	refx.Must(err)
	return func(options interface{}) (*WrapDriver, error) {
		result, err := constructor.Call(options)
		if err != nil {
			return nil, errors.WithMessage(err, "constructor.Call failed")
		}

		if constructor.ReturnError {
			if !result[1].IsNil() {
				return nil, errors.Wrapf(result[1].Interface().(error), "New failed")
			}
			return &WrapDriver{
				inner:      result[0],
				methodName: methodName,
			}, nil
		}

		return &WrapDriver{
			inner:      result[0],
			methodName: methodName,
		}, nil
	}
}

func (d *WrapDriver) Do(v interface{}) (interface{}, error) {
	methodName := d.methodName
	if methodName == "" {
		methodNameV, err := refx.InterfaceGet(v, d.methodKey)
		if err != nil {
			return nil, errors.Wrap(err, "refx.InterfaceGet failed")
		}
		var ok bool
		methodName, ok = methodNameV.(string)
		if !ok {
			return nil, NewError(nil, "driver.InvalidMethodName", "method should be string")
		}
	}
	method := d.inner.MethodByName(methodName)
	if !method.IsValid() {
		return nil, NewErrorf(nil, "driver.NoSuchMethod", "no such method [%s]", methodName)
	}

	mt := method.Type()
	if mt.NumIn() == 0 {
		// func()
		return resultToInterface(mt, method.Call([]reflect.Value{}))
	}
	if mt.NumIn() == 1 {
		// func(ctx)
		if mt.In(0) == reflect.TypeOf((*context.Context)(nil)).Elem() {
			return resultToInterface(mt, method.Call([]reflect.Value{reflect.ValueOf(context.Background())}))
		}

		// func(req)
		req := reflect.New(mt.In(0))
		err := refx.InterfaceToStruct(v, req.Interface())
		if err != nil {
			return nil, NewErrorf(err, "driver.ConstructReqFailed", "refx.InterfaceToStruct failed, err: [%s]", err.Error())
		}
		return resultToInterface(mt, method.Call([]reflect.Value{req.Elem()}))
	}
	if mt.NumIn() == 2 {
		// func(ctx, req)
		if mt.In(0) == reflect.TypeOf((*context.Context)(nil)).Elem() {
			req := reflect.New(mt.In(1))
			err := refx.InterfaceToStruct(v, req.Interface())
			if err != nil {
				return nil, NewErrorf(err, "driver.ConstructReqFailed", "refx.InterfaceToStruct failed, err: [%s]", err.Error())
			}
			return resultToInterface(mt, method.Call([]reflect.Value{reflect.ValueOf(context.Background()), req.Elem()}))
		}
	}

	var args []reflect.Value
	idx := 0
	// func(ctx, arg1, arg2, ...)
	if mt.In(0) == reflect.TypeOf((*context.Context)(nil)).Elem() {
		args = append(args, reflect.ValueOf(context.Background()))
		idx++
	}
	// func(arg1, arg2, ...)
	argsV, err := refx.InterfaceGet(v, "args")
	if err != nil {
		return nil, NewError(nil, "driver.MissingArgument.Args", "missing required field args")
	}
	argsRv := reflect.ValueOf(argsV)
	if argsRv.Type().Kind() != reflect.Slice {
		return nil, NewError(nil, "driver.InvalidArgument.Args", "args should be slice")
	}
	for i := 0; i < argsRv.Len(); i++ {
		arg := reflect.New(mt.In(idx))
		err := refx.InterfaceToStruct(argsRv.Index(i).Interface(), arg.Interface())
		if err != nil {
			return nil, NewErrorf(err, "driver.ConstructReqFailed", "refx.InterfaceToStruct failed, err: [%s]", err.Error())
		}
		args = append(args, arg.Elem())
		idx++
	}

	return resultToInterface(mt, method.Call(args))
}

func resultToInterface(mt reflect.Type, results []reflect.Value) (interface{}, error) {
	if mt.NumOut() == 0 {
		return map[string]interface{}{}, nil
	}
	if mt.NumOut() == 1 {
		if mt.Out(0) == reflect.TypeOf((*error)(nil)).Elem() {
			return nil, results[0].Interface().(error)
		}
		v, err := structToInterface(results[0].Interface())
		if err != nil {
			return nil, err
		}
		return v, nil
	}
	if mt.NumOut() == 2 {
		if mt.Out(1) != reflect.TypeOf((*error)(nil)).Elem() {
			return nil, NewError(nil, "driver.InvalidMethod", "the second result should be error")
		}
		if results[1].IsNil() {
			v, err := structToInterface(results[0].Interface())
			if err != nil {
				return nil, err
			}
			return v, nil
		}
		return nil, results[1].Interface().(error)
	}
	return nil, NewError(nil, "driver.InvalidMethod", "return too many values")
}

func structToInterface(src interface{}) (interface{}, error) {
	buf, err := jsoniter.Marshal(src)
	if err != nil {
		return nil, errors.Wrap(err, "jsoniter.Marshal failed")
	}
	var dst interface{}
	err = jsoniter.Unmarshal(buf, &dst)
	return dst, err
}
