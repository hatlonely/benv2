package eval

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/PaesslerAG/gval"
	"github.com/generikvault/gvalstrings"
	"github.com/hatlonely/go-kit/cast"
	uuid "github.com/satori/go.uuid"
)

var Lang = gval.NewLanguage(
	gval.Arithmetic(),
	gval.Bitmask(),
	gval.Text(),
	gval.PropositionalLogic(),
	gval.JSON(),
	gvalstrings.SingleQuoted(),
	gval.Function("date", func(arguments ...interface{}) (interface{}, error) {
		if len(arguments) != 1 {
			return nil, fmt.Errorf("date() expects exactly one string argument")
		}
		s, ok := arguments[0].(string)
		if !ok {
			return nil, fmt.Errorf("date() expects exactly one string argument")
		}
		for _, format := range [...]string{
			time.ANSIC,
			time.UnixDate,
			time.RubyDate,
			time.Kitchen,
			time.RFC3339,
			time.RFC3339Nano,
			"2006-01-02",                         // RFC 3339
			"2006-01-02 15:04",                   // RFC 3339 with minutes
			"2006-01-02 15:04:05",                // RFC 3339 with seconds
			"2006-01-02 15:04:05-07:00",          // RFC 3339 with seconds and timezone
			"2006-01-02T15Z0700",                 // ISO8601 with hour
			"2006-01-02T15:04Z0700",              // ISO8601 with minutes
			"2006-01-02T15:04:05Z0700",           // ISO8601 with seconds
			"2006-01-02T15:04:05.999999999Z0700", // ISO8601 with nanoseconds
		} {
			ret, err := time.ParseInLocation(format, s, time.Local)
			if err == nil {
				return ret, nil
			}
		}
		return nil, fmt.Errorf("date() could not parse %s", s)
	}),
	gval.Function("len", func(x interface{}) (int, error) {
		return len(x.(string)), nil
	}),
	gval.Function("int", func(x interface{}) (int, error) {
		switch v := x.(type) {
		case string:
			return cast.ToIntE(strings.Fields(v)[0])
		}
		return cast.ToIntE(x)
	}),
	gval.Function("uuid", func() (string, error) {
		return uuid.NewV4().String(), nil
	}),
	gval.Function("random", func() (float64, error) {
		return rand.Float64(), nil
	}),
	gval.Function("randInt", func(n int64) (int64, error) {
		return rand.Int63n(n), nil
	}),
)
