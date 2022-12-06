package driver

import (
	"context"
	"errors"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func NewAWithOptions(options *AOptions) (*A, error) {
	return &A{
		F1: options.F1,
		F2: options.F2,
	}, nil
}

type AOptions struct {
	F1 string
	F2 int
}

type A struct {
	F1 string
	F2 int
}

type Func0Res struct {
	ResF1 string
	ResF2 int
}

func (a *A) Func0() *Func0Res {
	return &Func0Res{
		ResF1: "val1",
		ResF2: 2,
	}
}

func (a *A) Func0Err() (*Func0Res, error) {
	return &Func0Res{
		ResF1: "val1",
		ResF2: 2,
	}, nil
}

type Func1Req struct {
	ReqF1 string
	ReqF2 int
}

type Func1Res struct {
	ResF1 string
	ResF2 int
}

func (a *A) Func1Err(req *Func1Req) (*Func1Res, error) {
	return &Func1Res{
		ResF1: req.ReqF1,
		ResF2: req.ReqF2,
	}, nil
}

type Func2Req struct {
	ReqF1 string
	ReqF2 int
}

type Func2Res struct {
	ResF1 string
	ResF2 int
}

func (a *A) Func2CtxErr(ctx context.Context, req *Func2Req) (*Func2Res, error) {
	return &Func2Res{
		ResF1: req.ReqF1 + req.ReqF1,
		ResF2: req.ReqF2 + req.ReqF2,
	}, nil
}

func (a *A) Func2CtxErrCode(ctx context.Context, req *Func2Req) (*Func2Res, error) {
	return nil, NewError(errors.New("timeout"), "Timeout", "Message")
}

type Func3Res struct {
	A int
	B int
	C string
}

func (a *A) Func3Err(a_ int, b int, c string) (*Func3Res, error) {
	return &Func3Res{
		A: a_,
		B: b,
		C: c,
	}, nil
}

func (a *A) Func3CtxErr(ctx context.Context, a_ int, b int, c string) (*Func3Res, error) {
	return &Func3Res{
		A: a_,
		B: b,
		C: c,
	}, nil
}

func TestWrapDriver(t *testing.T) {
	Convey("TestWrapDriver", t, func() {
		d, err := NewWrapDriverWithMethodKey(NewAWithOptions, "Method")(&AOptions{
			F1: "val1",
			F2: 2,
		})
		So(err, ShouldBeNil)
		So(d.inner.Interface(), ShouldResemble, &A{
			F1: "val1",
			F2: 2,
		})
		So(d.methodKey, ShouldEqual, "Method")

		Convey("Func0", func() {
			res, err := d.Do(map[string]interface{}{
				"Method": "Func0",
			})
			So(err, ShouldBeNil)
			So(res, ShouldResemble, map[string]interface{}{
				"ResF1": "val1",
				"ResF2": int64(2),
			})
		})

		Convey("Func0Err", func() {
			res, err := d.Do(map[string]interface{}{
				"Method": "Func0Err",
			})
			So(err, ShouldBeNil)
			So(res, ShouldResemble, map[string]interface{}{
				"ResF1": "val1",
				"ResF2": int64(2),
			})
		})

		Convey("Func1Err", func() {
			res, err := d.Do(map[string]interface{}{
				"Method": "Func1Err",
				"ReqF1":  "val1",
				"ReqF2":  1669914649166991464,
			})
			So(err, ShouldBeNil)

			So(res, ShouldResemble, map[string]interface{}{
				"ResF1": "val1",
				"ResF2": int64(1669914649166991464),
			})
		})

		Convey("Func2CtxErr", func() {
			res, err := d.Do(map[string]interface{}{
				"Method": "Func2CtxErr",
				"ReqF1":  "val2",
				"ReqF2":  22,
			})
			So(err, ShouldBeNil)

			So(res, ShouldResemble, map[string]interface{}{
				"ResF1": "val2val2",
				"ResF2": int64(44),
			})
		})

		Convey("Func3Err", func() {
			res, err := d.Do(map[string]interface{}{
				"Method": "Func3Err",
				"args": []interface{}{
					1,
					"2",
					"val3",
				},
			})
			So(err, ShouldBeNil)

			So(res, ShouldResemble, map[string]interface{}{
				"A": int64(1),
				"B": int64(2),
				"C": "val3",
			})
		})

		Convey("Func3CtxErr", func() {
			res, err := d.Do(map[string]interface{}{
				"Method": "Func3CtxErr",
				"args": []interface{}{
					1,
					"2",
					"val3",
				},
			})
			So(err, ShouldBeNil)

			So(res, ShouldResemble, map[string]interface{}{
				"A": int64(1),
				"B": int64(2),
				"C": "val3",
			})
		})

		Convey("error", func() {
			_, err := d.Do(map[string]interface{}{
				"Method": "Func2CtxErrCode",
				"ReqF1":  "val2",
				"ReqF2":  22,
			})
			So(err, ShouldNotBeNil)
			So(err.(*Error).Code, ShouldEqual, "Timeout")
			So(err.(*Error).Message, ShouldEqual, "Message")
		})
	})

}
