package driver

//import (
//	"github.com/hatlonely/go-kit/refx"
//	"github.com/hatlonely/go-kit/wrap"
//	"github.com/pkg/errors"
//)
//
//type OSSDriver struct {
//	inner interface{}
//}
//
//func NewOSSDriverWithOptions(options *wrap.OSSClientWrapperOptions) (*OSSDriver, error) {
//	v, err := wrap.NewOSSClientWrapperWithOptions(options)
//	if err != nil {
//		return nil, errors.WithMessage(err, "wrap.NewOSSClientWrapperWithOptions failed")
//	}
//
//	return &OSSDriver{v}, nil
//}
//
//func (d *OSSDriver) Do(v interface{}) (interface{}, error) {
//	method, err := refx.InterfaceGet(v, "Action")
//	if err != nil {
//		return nil, errors.Wrap(err, "refx.InterfaceGet failed")
//	}
//
//
//}
