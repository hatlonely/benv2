package recorder

import (
	"time"

	"github.com/pkg/errors"
)

type StatisticsOptions struct {
	PointNumber int `dft:"100"`
	Interval    time.Duration
}

type Statistics struct {
	options *StatisticsOptions
}

type Measurement struct {
	Time               time.Time
	ResTimeMs          time.Duration
	Total              int
	Pass               int
	Fail               int
	QPS                int
	AvgResTimeMs       time.Duration
	SuccessRatePercent float64
}

func (s *Statistics) Statistics(analyst Analyst) ([]*Measurement, error) {
	st, et, err := analyst.TimeRange()
	et.Add(1) // 边界处理
	if err != nil {
		return nil, errors.WithMessage(err, "analyst.TimeRange")
	}
	interval := s.options.Interval
	if s.options.Interval == 0 {
		if s.options.PointNumber == 0 {
			s.options.PointNumber = 100
		}
		interval = et.Sub(st) / time.Duration(s.options.PointNumber)
	}

	var measurements []*Measurement
	for i := st; i.Before(et); i.Add(interval) {
		measurements = append(measurements, &Measurement{
			Time: i,
		})
	}

	stream, err := analyst.UnitStatStream()
	if err != nil {
		return nil, errors.WithMessage(err, "analyst.UnitStatStream failed")
	}
	for {
		stat, err := stream.Next()
		if err != nil {
			return nil, errors.WithMessage(err, "stream.Next failed")
		}

		t, err := time.Parse(time.RFC3339Nano, stat.Time)
		if err != nil {
			return nil, errors.WithMessage(err, "time.Parse failed")
		}
		idx := t.Sub(st) / interval

		measurement := measurements[idx]
		measurement.ResTimeMs += stat.ResTime
		measurement.Total += 1

		if stat == nil {
			break
		}
	}

	return measurements, nil
}
