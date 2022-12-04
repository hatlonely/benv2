package framework

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/hatlonely/benv2/internal/driver"
	"github.com/hatlonely/benv2/internal/eval"
	"github.com/hatlonely/benv2/internal/recorder"
	"github.com/hatlonely/benv2/internal/source"
	"github.com/hatlonely/go-kit/refx"
	"github.com/pkg/errors"
)

type Options struct {
	Name     string
	Recorder refx.TypeOptions
	Ctx      map[string]refx.TypeOptions
	Source   map[string]refx.TypeOptions
	Plan     struct {
		Duration time.Duration
		Unit     []struct {
			Parallel int `dft:"1"`
			Name     string
			Step     []*struct {
				Ctx string
				Req interface{}
			}
		}
	}
}

func NewFrameworkWithOptions(options *Options, opts ...refx.Option) (*Framework, error) {
	var err error
	fw := &Framework{
		name:   options.Name,
		ctx:    map[string]driver.Driver{},
		source: map[string]source.Source{},
	}

	fw.recorder, err = recorder.NewRecorderWithOptions(&options.Recorder, opts...)
	if err != nil {
		return nil, errors.WithMessage(err, "recorder.NewRecorderWithOptions failed")
	}

	for key, refxOptions := range options.Ctx {
		fw.ctx[key], err = driver.NewDriverWithOptions(&refxOptions, opts...)
		if err != nil {
			return nil, errors.WithMessage(err, "driver.NewDriverWithOptions failed")
		}
	}

	for key, refxOptions := range options.Source {
		fw.source[key], err = source.NewSourceWithOptions(&refxOptions, opts...)
		if err != nil {
			return nil, errors.WithMessage(err, "source.NewSourceWithOptions failed")
		}
	}

	var plan PlanInfo
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
			Parallel: unitDesc.Parallel,
			Name:     unitDesc.Name,
			Step:     step,
		})
	}

	fw.plan = &plan

	return fw, nil
}

type Framework struct {
	name     string
	recorder recorder.Recorder
	ctx      map[string]driver.Driver
	source   map[string]source.Source
	plan     *PlanInfo
}

type PlanInfo struct {
	Duration time.Duration
	Unit     []*UnitInfo
}

type UnitInfo struct {
	Parallel int
	Name     string
	Step     []*StepInfo
}

type StepInfo struct {
	Ctx string
	Req *eval.Evaluable
}

//type UnitStat struct {
//	Time    string
//	Name    string
//	Step    []*StepStat
//	ErrCode string
//	ResTime time.Duration
//}
//
//type StepStat struct {
//	Time    string
//	Req     interface{}
//	Res     interface{}
//	Err     error
//	ErrCode string
//	ResTime time.Duration
//}

func (fw *Framework) Run() error {
	var wg sync.WaitGroup
	defer fw.recorder.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	for _, unit := range fw.plan.Unit {
		for i := 0; i < unit.Parallel; i++ {
			wg.Add(1)
			go func(unit *UnitInfo) {
			out:
				for {
					select {
					case <-ctx.Done():
						break out
					case <-time.After(fw.plan.Duration):
						break out
					default:
						stat, err := fw.RunUnit(unit)
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
			}(unit)
		}
	}

	wg.Wait()

	return nil
}

func (fw *Framework) RunUnit(info *UnitInfo) (*recorder.UnitStat, error) {
	stat := &recorder.UnitStat{Time: time.Now().Format(time.RFC3339Nano), Name: info.Name}
	var err error

	// fetch source
	sourceMap := map[string]interface{}{}
	for key, src := range fw.source {
		sourceMap[key] = src.Fetch()
	}

	var req interface{}
	var stepResTime time.Duration

	unitStart := time.Now()
	var stepStart time.Time
	for _, step := range info.Step {
		req, err = step.Req.Evaluate(map[string]interface{}{
			"source": sourceMap,
			"stat":   stat,
		})
		if err != nil {
			return nil, errors.WithMessage(err, "step.Req.Evaluate failed")
		}
		d, ok := fw.ctx[step.Ctx]
		if !ok {
			return nil, errors.Errorf("ctx not found. ctx: [%s]", step.Ctx)
		}

		stepStart = time.Now()
		res, err := d.Do(req)
		stepResTime = time.Since(stepStart)
		if err != nil {
			err = errors.WithMessage(err, "driver.Do failed")
			break
		}

		stat.Step = append(stat.Step, &recorder.StepStat{
			Time:    stepStart.Format(time.RFC3339Nano),
			Req:     req,
			Res:     res,
			Err:     nil,
			ResTime: stepResTime,
		})
	}

	if err != nil {
		stat.Step = append(stat.Step, &recorder.StepStat{
			Time:    stepStart.Format(time.RFC3339Nano),
			Req:     req,
			Res:     nil,
			Err:     err,
			ResTime: stepResTime,
		})
	}

	stat.ResTime = time.Since(unitStart)

	return stat, nil
}
