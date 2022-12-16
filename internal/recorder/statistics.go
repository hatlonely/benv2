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
	Summary             map[string]*Summary
	QPS                 map[string][]*Measurement
	AvgResTimeMs        map[string][]*Measurement
	SuccessRatePercent  map[string][]*Measurement
	ErrCodeDistribution map[string]map[string]int
}

type Measurement struct {
	Time  time.Time
	Value float64
}

type Summary struct {
	Total              int
	Pass               int
	QPS                float64
	AvgResTimeMs       float64
	SuccessRatePercent float64
}

func (s *Statistics) Statistics(id string, analyst Analyst) ([]*Metric, error) {
	aggregations, err := s.aggregation(id, analyst)
	if err != nil {
		return nil, errors.WithMessage(err, "aggregation failed")
	}

	var metrics []*Metric
	for _, aggregationMap := range aggregations {
		metric, err := s.calculate(aggregationMap)
		if err != nil {
			return nil, errors.WithMessage(err, "s.calculate failed")
		}
		metrics = append(metrics, metric)
	}

	return metrics, nil
}

func (s *Statistics) calculate(aggregationMap map[string][]*Aggregation) (*Metric, error) {
	summaryMap := map[string]*Summary{}
	qpsMap := map[string][]*Measurement{}
	avgResTimeMsMap := map[string][]*Measurement{}
	successRatePercentMap := map[string][]*Measurement{}
	errCodeDistributionMap := map[string]map[string]int{}

	for key, aggregations := range aggregationMap {
		var summary Summary
		totalResTime := time.Duration(0)
		// 丢弃最后一次结果
		for i := 0; i < len(aggregations)-1; i++ {
			summary.Total += aggregations[i].Total
			summary.Pass += aggregations[i].Pass
			totalResTime += aggregations[i].PassResTime
		}
		summary.QPS = float64(summary.Pass) / aggregations[len(aggregations)-1].Time.Sub(aggregations[0].Time).Seconds()
		summary.AvgResTimeMs = float64(totalResTime.Milliseconds()) / float64(summary.Pass)
		summary.SuccessRatePercent = float64(summary.Pass*100) / float64(summary.Total)
		summaryMap[key] = &summary

		qps := make([]*Measurement, 0, len(aggregations))
		for _, aggregation := range aggregations {
			qps = append(qps, &Measurement{
				Time:  aggregation.Time,
				Value: float64(aggregation.Pass) / aggregation.Duration.Seconds(),
			})
		}
		qpsMap[key] = qps

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
		avgResTimeMsMap[key] = avgResTimeMs

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
		successRatePercentMap[key] = successRatePercent

		errCodeDistribution := map[string]int{}
		for _, aggregation := range aggregations {
			for key, val := range aggregation.ErrCode {
				errCodeDistribution[key] += val
			}
		}
		errCodeDistributionMap[key] = errCodeDistribution
	}

	return &Metric{
		QPS:                 qpsMap,
		AvgResTimeMs:        avgResTimeMsMap,
		SuccessRatePercent:  successRatePercentMap,
		ErrCodeDistribution: errCodeDistributionMap,
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

func (s *Statistics) aggregation(id string, analyst Analyst) ([]map[string][]*Aggregation, error) {
	meta, err := analyst.Meta()
	if err != nil {
		return nil, errors.WithMessage(err, "analyst.TimeRange")
	}

	interval := s.options.Interval
	if s.options.Interval == 0 {
		if s.options.PointNumber == 0 {
			s.options.PointNumber = 100
		}
		interval = meta.Duration / time.Duration(s.options.PointNumber)
	}

	aggregationIdxMap := map[int]map[string][]*Aggregation{}

	stream, err := analyst.UnitStatStream(id)
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

		timeRange := meta.TimeRange[stat.Seq]
		idx := int(t.Sub(timeRange.StartTime) / interval)

		aggregationMap, ok := aggregationIdxMap[stat.Seq]
		if !ok {
			aggregationMap = map[string][]*Aggregation{}
			aggregationIdxMap[stat.Seq] = aggregationMap
		}
		if _, ok := aggregationMap[stat.Name]; !ok {
			var aggregations []*Aggregation
			for i := timeRange.StartTime; i.Before(timeRange.EndTime.Add(interval)); i = i.Add(interval) {
				aggregations = append(aggregations, &Aggregation{
					Time:     i,
					Duration: interval,
					ErrCode:  map[string]int{},
				})
			}
			aggregationMap[stat.Name] = aggregations
		}
		if idx >= len(aggregationMap[stat.Name]) {
			idx = len(aggregationMap[stat.Name])
		}

		aggregation := aggregationMap[stat.Name][idx]
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

	var aggregations []map[string][]*Aggregation
	for i := 0; i < len(aggregationIdxMap); i++ {
		aggregations = append(aggregations, aggregationIdxMap[i])
	}

	return aggregations, nil
}
