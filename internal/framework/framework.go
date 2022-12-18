package framework

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/PaesslerAG/gval"
	"github.com/hatlonely/go-kit/refx"
	"github.com/pkg/errors"

	"github.com/hatlonely/benv2/internal/driver"
	"github.com/hatlonely/benv2/internal/eval"
	"github.com/hatlonely/benv2/internal/monitor"
	"github.com/hatlonely/benv2/internal/recorder"
	"github.com/hatlonely/benv2/internal/reporter"
	"github.com/hatlonely/benv2/internal/source"
)

type Options struct {
	ID     string
	Name   string
	Ctx    map[string]refx.TypeOptions
	Source map[string]refx.TypeOptions
	Plan   struct {
		Duration time.Duration
		Interval time.Duration
		Parallel []map[string]int
		Unit     []struct {
			Name string
			Step []*struct {
				Ctx     string
				Req     interface{}
				ErrCode string
				Success string
			}
		}
	}
	Recorder   refx.TypeOptions
	Analyst    refx.TypeOptions
	Statistics recorder.StatisticsOptions
	Monitors   []refx.TypeOptions
	Reporter   refx.TypeOptions
}

func NewFrameworkWithOptions(options *Options, opts ...refx.Option) (*Framework, error) {
	var err error
	ctx := map[string]driver.Driver{}
	for key, refxOptions := range options.Ctx {
		ctx[key], err = driver.NewDriverWithOptions(&refxOptions, opts...)
		if err != nil {
			return nil, errors.WithMessage(err, "driver.NewDriverWithOptions failed")
		}
	}

	source_ := map[string]source.Source{}
	for key, refxOptions := range options.Source {
		source_[key], err = source.NewSourceWithOptions(&refxOptions, opts...)
		if err != nil {
			return nil, errors.WithMessage(err, "source.NewSourceWithOptions failed")
		}
	}

	plan := &PlanInfo{
		Duration: options.Plan.Duration,
		Interval: options.Plan.Interval,
		Parallel: options.Plan.Parallel,
	}
	for _, unitDesc := range options.Plan.Unit {
		var step []*StepInfo
		for _, stepDesc := range unitDesc.Step {
			reqEval, err := eval.NewEvaluable(stepDesc.Req)
			if err != nil {
				return nil, errors.WithMessage(err, "eval.NewEvaluable failed")
			}
			var errCodeEval gval.Evaluable
			if stepDesc.ErrCode != "" {
				errCodeEval, err = eval.Lang.NewEvaluable(stepDesc.ErrCode)
				if err != nil {
					return nil, errors.WithMessage(err, "eval.NewEvaluable failed")
				}
			}
			var successEval gval.Evaluable
			if stepDesc.Success != "" {
				successEval, err = eval.Lang.NewEvaluable(stepDesc.Success)
				if err != nil {
					return nil, errors.WithMessage(err, "eval.NewEvaluable failed")
				}
			}
			step = append(step, &StepInfo{
				Ctx:     stepDesc.Ctx,
				Req:     reqEval,
				ErrCode: errCodeEval,
				Success: successEval,
			})
		}
		plan.Unit = append(plan.Unit, &UnitInfo{
			Name: unitDesc.Name,
			Step: step,
		})
	}

	recorder_, err := recorder.NewRecorderWithOptions(&options.Recorder, opts...)
	if err != nil {
		return nil, errors.WithMessage(err, "recorder.NewRecorderWithOptions failed")
	}

	var analyst recorder.Analyst
	if options.Analyst.Type != "" {
		analyst, err = recorder.NewAnalystWithOptions(&options.Analyst, opts...)
		if err != nil {
			return nil, errors.WithMessage(err, "recorder.NewAnalystWithOptions failed")
		}
	}

	statistics := recorder.NewStatisticsWithOptions(&options.Statistics)

	if options.Reporter.Type == "" {
		options.Reporter.Type = "Json"
	}
	reporter_, err := reporter.NewReporterWithOptions(&options.Reporter, opts...)
	if err != nil {
		return nil, errors.WithMessage(err, "reporter.NewReporterWithOptions failed")
	}

	var monitors []monitor.Monitor
	for _, opt := range options.Monitors {
		monitor_, err := monitor.NewMonitorWithOptions(&opt, opts...)
		if err != nil {
			return nil, errors.WithMessage(err, "monitor.NewMonitorWithOptions failed")
		}
		monitors = append(monitors, monitor_)
	}

	return &Framework{
		id:         options.ID,
		name:       options.Name,
		ctx:        ctx,
		source:     source_,
		plan:       plan,
		recorder:   recorder_,
		analyst:    analyst,
		statistics: statistics,
		monitors:   monitors,
		reporter:   reporter_,
	}, nil
}

type Framework struct {
	id         string
	name       string
	ctx        map[string]driver.Driver
	source     map[string]source.Source
	plan       *PlanInfo
	recorder   recorder.Recorder
	analyst    recorder.Analyst
	statistics *recorder.Statistics
	monitors   []monitor.Monitor
	reporter   reporter.Reporter
}

type PlanInfo struct {
	Duration time.Duration
	Interval time.Duration
	Parallel []map[string]int
	Unit     []*UnitInfo
}

type UnitInfo struct {
	Name string
	Step []*StepInfo
}

type StepInfo struct {
	Ctx     string
	Req     *eval.Evaluable
	ErrCode gval.Evaluable
	Success gval.Evaluable
}

func (fw *Framework) Run() error {
	meta := &recorder.Meta{
		ID:       fw.id,
		Name:     fw.name,
		Parallel: fw.plan.Parallel,
		Duration: fw.plan.Duration,
	}

	startTime := time.Now().Add(time.Second)

	for idx, parallelMap := range fw.plan.Parallel {
		meta.TimeRange = append(meta.TimeRange, &recorder.TimeRange{
			StartTime: startTime,
			EndTime:   startTime.Add(fw.plan.Duration),
		})

		var wg sync.WaitGroup
		ctx, cancel := context.WithCancel(context.Background())
		for _, unit := range fw.plan.Unit {
			parallel, ok := parallelMap[unit.Name]
			if !ok {
				continue
			}
			for i := 0; i < parallel; i++ {
				wg.Add(1)
				go func(unit *UnitInfo, idx int) {
					time.Sleep(time.Until(startTime))
					deadline := time.After(fw.plan.Duration)
				out:
					for {
						select {
						case <-ctx.Done():
							break out
						case <-deadline:
							break out
						default:
							stat, err := fw.RunUnit(unit)
							if err != nil {
								fmt.Println(err)
								cancel()
								break
							}
							stat.Seq = idx
							err = fw.recorder.Record(stat)
							if err != nil {
								fmt.Println(err)
								cancel()
								break
							}
						}
					}
					wg.Done()
				}(unit, idx)
			}
		}
		wg.Wait()
		cancel()

		startTime = startTime.Add(fw.plan.Duration + fw.plan.Interval)
	}

	if err := fw.recorder.RecordMeta(meta); err != nil {
		return errors.WithMessage(err, "recorder.RecordMeta failed")
	}

	_ = fw.recorder.Close()

	if fw.analyst != nil {
		metrics, err := fw.statistics.Statistics(fw.id, fw.analyst)
		if err != nil {
			return errors.WithMessage(err, "statistics.Statistics failed")
		}

		meta, err := fw.analyst.Meta()
		if err != nil {
			return errors.WithMessage(err, "analyst.Meta failed")
		}

		var measurementMapSlice []map[string]map[string][]*recorder.Measurement
		for _, timeRange := range meta.TimeRange {
			measurementMap := map[string]map[string][]*recorder.Measurement{}
			for _, monitor_ := range fw.monitors {
				//mm, err := monitor_.Collect(timeRange.StartTime, timeRange.EndTime)
				_ = timeRange
				mm, err := monitor_.Collect(time.Now().Add(-10*time.Minute), time.Now())
				if err != nil {
					return errors.WithMessage(err, "monitor.Collect failed")
				}
				for key, val := range mm {
					measurementMap[key] = val
				}
			}
			measurementMapSlice = append(measurementMapSlice, measurementMap)
		}

		fmt.Println(fw.reporter.Report(meta, metrics, measurementMapSlice))
	}

	return nil
}

func (fw *Framework) RunUnit(info *UnitInfo) (*recorder.UnitStat, error) {
	unitStat := &recorder.UnitStat{Name: info.Name, ID: fw.id}
	var err error

	// fetch source
	sourceMap := map[string]interface{}{}
	for key, src := range fw.source {
		sourceMap[key] = src.Fetch()
	}

	var req interface{}
	var stepResTime time.Duration

	unitStart := time.Now()
	for _, step := range info.Step {
		req, err = step.Req.Evaluate(map[string]interface{}{
			"source": sourceMap,
			"stat":   unitStat,
		})
		if err != nil {
			return nil, errors.WithMessage(err, "step.Req.Evaluate failed")
		}
		d, ok := fw.ctx[step.Ctx]
		if !ok {
			return nil, errors.Errorf("ctx not found. ctx: [%s]", step.Ctx)
		}

		stepStart := time.Now()
		var res interface{}
		res, err = d.Do(req)
		stepResTime = time.Since(stepStart)
		if err != nil {
			err = errors.WithMessage(err, "driver.Do failed")
			break
		}

		errCode := ""
		if step.Success != nil {
			success, e := step.Success.EvalBool(context.Background(), map[string]interface{}{
				"res": res,
			})
			if e != nil {
				return nil, errors.WithMessage(e, "step.Success.Evaluate failed")
			}
			if !success {
				if step.ErrCode == nil {
					errCode = "Fail"
				} else {
					errCodeV, e := step.ErrCode(context.Background(), map[string]interface{}{
						"res": res,
					})
					if e != nil {
						return nil, errors.WithMessage(e, "step.ErrCode.Evaluate failed")
					}
					errCode = fmt.Sprintf("%v", errCodeV)
				}
			}
		}

		unitStat.Step = append(unitStat.Step, &recorder.StepStat{
			Time:    time.Now().Format(time.RFC3339Nano),
			Req:     req,
			Res:     res,
			Err:     nil,
			ResTime: stepResTime,
			ErrCode: errCode,
		})

		if errCode != "" {
			unitStat.ErrCode = errCode
			unitStat.ResTime = time.Since(unitStart)
			unitStat.Time = time.Now().Format(time.RFC3339Nano)
			return unitStat, nil
		}
	}

	if err != nil {
		stepStat := &recorder.StepStat{
			Time:    time.Now().Format(time.RFC3339Nano),
			Req:     req,
			Res:     nil,
			Err:     err,
			ResTime: stepResTime,
			ErrCode: "Internal",
		}

		switch e := errors.Cause(err).(type) {
		case *driver.Error:
			stepStat.ErrCode = e.Code
		}

		unitStat.Step = append(unitStat.Step, stepStat)
		unitStat.ErrCode = stepStat.ErrCode
	}

	unitStat.ResTime = time.Since(unitStart)
	unitStat.Time = time.Now().Format(time.RFC3339Nano)

	return unitStat, nil
}
