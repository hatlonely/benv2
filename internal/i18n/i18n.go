package i18n

import (
	"reflect"

	"github.com/pkg/errors"
)

type Title struct {
	Report              string
	Summary             string
	QPS                 string
	AvgResTimeMs        string
	SuccessRatePercent  string
	ErrCodeDistribution string
}

type Tooltip struct {
	Save string
	Copy string
}

type I18n struct {
	Title   Title
	Tooltip Tooltip
}

var defaultI18n = map[string]*I18n{
	"dft": &I18n{
		Title: Title{
			Report:              "Report",
			Summary:             "Summary",
			QPS:                 "QPS",
			AvgResTimeMs:        "AvgResTimeMs",
			SuccessRatePercent:  "SuccessRatePercent",
			ErrCodeDistribution: "ErrCodeDistribution",
		},
		Tooltip: Tooltip{
			Save: "Save",
			Copy: "Copy",
		},
	},
}

type Options struct {
	Lang string `dft:"dft"`
	I18n I18n
}

func NewI18nWithOptions(options *Options) (*I18n, error) {
	if options.Lang == "" {
		options.Lang = "dft"
	}

	dftI18n, ok := defaultI18n[options.Lang]
	if !ok {
		return nil, errors.Errorf("unknown lang [%s]", options.Lang)
	}

	i18n := options.I18n
	merge(&i18n, dftI18n)

	return &i18n, nil
}

func merge(i18n interface{}, dftI18n interface{}) {
	rv := reflect.ValueOf(i18n).Elem()
	drv := reflect.ValueOf(dftI18n).Elem()

	for i := 0; i < rv.NumField(); i++ {
		if rv.Field(i).Type().Kind() == reflect.Struct {
			merge(rv.Field(i).Addr().Interface(), drv.Field(i).Addr().Interface())
			continue
		}
		if rv.Field(i).IsZero() {
			rv.Field(i).Set(drv.Field(i))
		}
	}
}
