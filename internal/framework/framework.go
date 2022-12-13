package framework

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/hatlonely/benv2/internal/driver"
	"github.com/hatlonely/benv2/internal/eval"
	"github.com/hatlonely/benv2/internal/recorder"
	"github.com/hatlonely/benv2/internal/reporter"
	"github.com/hatlonely/benv2/internal/source"
	"github.com/hatlonely/go-kit/refx"
	"github.com/pkg/errors"
)

type Options struct {
	Name   string
	Ctx    map[string]refx.TypeOptions
	Source map[string]refx.TypeOptions
	Plan   struct {
		Duration time.Duration
		Parallel []map[string]int
		Unit     []struct {
			Name string
			Step []*struct {
				Ctx string
				Req interface{}
			}
		}
	}
	Recorder   refx.TypeOptions
	Analyst    refx.TypeOptions
	Statistics recorder.StatisticsOptions
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
		Parallel: options.Plan.Parallel,
	}
	for _, unitDesc := range options.Plan.Unit {
		var step []*StepInfo
		for _, stepDesc := range unitDesc.Step {
			e, err := eval.NewEvaluable(stepDesc.Req)
			if err != nil {
				return nil, errors.WithMessage(err, "eval.NewEvaluable failed")
			}
			step = append(step, &StepInfo{
				Ctx: stepDesc.Ctx,
				Req: e,
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
			return nil, errors.WithMessage(err, "recorder.NewRecorderWithOptions failed")
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

	return &Framework{
		name:       options.Name,
		ctx:        ctx,
		source:     source_,
		plan:       plan,
		recorder:   recorder_,
		analyst:    analyst,
		statistics: statistics,
		reporter:   reporter_,
	}, nil
}

type Framework struct {
	name       string
	ctx        map[string]driver.Driver
	source     map[string]source.Source
	plan       *PlanInfo
	recorder   recorder.Recorder
	analyst    recorder.Analyst
	statistics *recorder.Statistics
	reporter   reporter.Reporter
}

type PlanInfo struct {
	Duration time.Duration
	Parallel []map[string]int
	Unit     []*UnitInfo
}

type UnitInfo struct {
	Name string
	Step []*StepInfo
}

type StepInfo struct {
	Ctx string
	Req *eval.Evaluable
}

func (fw *Framework) Run() error {
	if err := fw.recorder.RecordMeta(&recorder.Meta{
		Name:     fw.name,
		Parallel: fw.plan.Parallel,
	}); err != nil {
		return errors.WithMessage(err, "recorder.RecordMeta failed")
	}

	for idx, parallelMap := range fw.plan.Parallel {
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
							stat.Seq = idx
							if err != nil {
								fmt.Println(err)
								cancel()
							}
							err = fw.recorder.Record(stat)
							if err != nil {
								fmt.Println(err)
								cancel()
							}
						}
					}
					wg.Done()
				}(unit, idx)
			}
		}
		wg.Wait()
		cancel()
	}

	fw.recorder.Close()

	if fw.analyst != nil {
		metrics, err := fw.statistics.Statistics(fw.analyst)
		if err != nil {
			return errors.WithMessage(err, "statistics.Statistics failed")
		}

		meta, err := fw.analyst.Meta()
		if err != nil {
			return errors.WithMessage(err, "analyst.Meta failed")
		}

		fmt.Println(fw.reporter.Report(meta, metrics))
	}

	return nil
}

func (fw *Framework) RunUnit(info *UnitInfo) (*recorder.UnitStat, error) {
	unitStat := &recorder.UnitStat{Name: info.Name}
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
		res, err := d.Do(req)
		stepResTime = time.Since(stepStart)
		if err != nil {
			err = errors.WithMessage(err, "driver.Do failed")
			break
		}

		unitStat.Step = append(unitStat.Step, &recorder.StepStat{
			Time:    time.Now().Format(time.RFC3339Nano),
			Req:     req,
			Res:     res,
			Err:     nil,
			ResTime: stepResTime,
		})
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
