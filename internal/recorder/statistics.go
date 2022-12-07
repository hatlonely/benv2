package recorder

import (
	"time"

	"github.com/pkg/errors"
)

func NewStatisticsWithOptions(options *StatisticsOptions) *Statistics {
	return &Statistics{
		options: options,
	}
}

type StatisticsOptions struct {
	PointNumber int `dft:"100"`
	Interval    time.Duration
}

type Statistics struct {
	options *StatisticsOptions
}

type Metric struct {
	QPS                 []*Measurement
	AvgResTimeMs        []*Measurement
	SuccessRatePercent  []*Measurement
	ErrCodeDistribution map[string]int
}
type Measurement struct {
	Time  time.Time
	Value float64
}

func (s *Statistics) Statistics(analyst Analyst) (*Metric, error) {
	aggregations, err := s.aggregation(analyst)
	if err != nil {
		return nil, errors.WithMessage(err, "aggregation failed")
	}

	qps := make([]*Measurement, 0, len(aggregations))
	for _, aggregation := range aggregations {
		qps = append(qps, &Measurement{
			Time:  aggregation.Time,
			Value: float64(aggregation.Pass) / aggregation.Duration.Seconds(),
		})
	}

	avgResTimeMs := make([]*Measurement, 0, len(aggregations))
	for _, aggregation := range aggregations {
		if aggregation.Pass == 0 {
			continue
		}
		avgResTimeMs = append(avgResTimeMs, &Measurement{
			Time:  aggregation.Time,
			Value: float64(aggregation.PassResTime.Milliseconds()) / float64(aggregation.Pass),
		})
	}

	successRatePercent := make([]*Measurement, 0, len(aggregations))
	for _, aggregation := range aggregations {
		if aggregation.Total == 0 {
			continue
		}
		successRatePercent = append(successRatePercent, &Measurement{
			Time:  aggregation.Time,
			Value: float64(aggregation.Pass*100) / float64(aggregation.Total),
		})
	}

	errCodeDistribution := map[string]int{}
	for _, aggregation := range aggregations {
		for key, val := range aggregation.ErrCode {
			errCodeDistribution[key] += val
		}
	}

	return &Metric{
		QPS:                 qps,
		AvgResTimeMs:        avgResTimeMs,
		SuccessRatePercent:  successRatePercent,
		ErrCodeDistribution: errCodeDistribution,
	}, nil
}

type Aggregation struct {
	Time         time.Time
	Duration     time.Duration
	Total        int
	TotalResTime time.Duration
	Pass         int
	PassResTime  time.Duration
	Fail         int
	ErrCode      map[string]int
}

func (s *Statistics) aggregation(analyst Analyst) ([]*Aggregation, error) {
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

	var aggregations []*Aggregation
	for i := st; i.Before(et); i.Add(interval) {
		aggregations = append(aggregations, &Aggregation{
			Time:     i,
			Duration: interval,
			ErrCode:  map[string]int{},
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

		if stat == nil {
			break
		}

		t, err := time.Parse(time.RFC3339Nano, stat.Time)
		if err != nil {
			return nil, errors.WithMessage(err, "time.Parse failed")
		}
		idx := t.Sub(st) / interval

		aggregation := aggregations[idx]
		aggregation.Total += 1
		aggregation.TotalResTime += stat.ResTime
		if stat.ErrCode != "" {
			aggregation.Fail += 1
			aggregation.ErrCode[stat.ErrCode] += 1
		} else {
			aggregation.Pass += 1
			aggregation.PassResTime += stat.ResTime
			aggregation.ErrCode["OK"] += 1
		}
	}

	return aggregations, nil
}
