package reporter

import (
	"bytes"
	"fmt"
	"math"
	"text/template"
	"time"

	"github.com/hatlonely/go-kit/strx"
	"github.com/pkg/errors"

	"github.com/hatlonely/benv2/internal/i18n"
	"github.com/hatlonely/benv2/internal/recorder"
)

type HtmlReporterOptions struct {
	Font struct {
		Style   string
		Body    string `dft:"'Roboto Condensed', sans-serif !important"`
		Code    string `dft:"'JetBrains Mono', monospace !important"`
		Echarts string `dft:"Roboto Condensed"`
	}
	Extra struct {
		Head       string
		BodyHeader string
		BodyFooter string
	}
	Padding struct {
		X int `dft:"2"`
		Y int `dft:"2"`
	}
	I18n i18n.Options
}

func NewHtmlReporterWithOptions(options *HtmlReporterOptions) (*HtmlReporter, error) {
	i18n_, err := i18n.NewI18nWithOptions(&options.I18n)
	if err != nil {
		return nil, errors.WithMessage(err, "i18n.NewI18nWithOptions failed")
	}

	reporter := &HtmlReporter{
		i18n:    i18n_,
		options: options,
	}

	funcs := template.FuncMap{
		"JsonMarshal":       strx.JsonMarshal,
		"JsonMarshalIndent": strx.JsonMarshalIndent,
		"RenderUnit":        reporter.RenderUnit,
		"RenderSummary":     reporter.RenderSummary,
		"FormatFloat": func(v float64) string {
			return fmt.Sprintf("%.2f", v)
		},
		"MeasurementToSerial": func(measurements []*recorder.Measurement) [][]interface{} {
			var items [][]interface{}
			for _, measurement := range measurements {
				items = append(items, []interface{}{
					measurement.Time.Format(time.RFC3339Nano), math.Round(measurement.Value*100) / 100,
				})
			}
			return items[:len(items)-1]
		},
		"EchartCodeRadius1": func(idx int, len int) int {
			return (70/len)*idx + 15
		},
		"EchartCodeRadius2": func(idx int, len int) int {
			return (70/len)*(idx+1) + 10
		},
		"DictToItems": func(d map[string]int) interface{} {
			var items []map[string]interface{}
			for k, v := range d {
				items = append(items, map[string]interface{}{
					"name":  k,
					"value": v,
				})
			}
			return items
		},
		"plusOne": func(i int) int {
			return i + 1
		},
	}

	reporter.reportTpl = template.Must(template.New("").Funcs(funcs).Parse(reportTplStr))
	reporter.summaryTpl = template.Must(template.New("").Funcs(funcs).Parse(summaryTplStr))
	reporter.unitTpl = template.Must(template.New("").Funcs(funcs).Parse(unitTplStr))

	return reporter, nil
}

type HtmlReporter struct {
	i18n    *i18n.I18n
	options *HtmlReporterOptions

	reportTpl  *template.Template
	summaryTpl *template.Template
	unitTpl    *template.Template
}

func (r *HtmlReporter) Report(meta *recorder.Meta, metrics []*recorder.Metric) string {
	var buf bytes.Buffer

	if err := r.reportTpl.Execute(&buf, map[string]interface{}{
		"Meta":      meta,
		"Customize": r.options,
		"I18n":      r.i18n,
		"Metrics":   metrics,
	}); err != nil {
		return fmt.Sprintf("%+v", errors.Wrap(err, "reportTpl.Execute failed"))
	}

	return buf.String()
}

func (r *HtmlReporter) RenderSummary(meta *recorder.Meta, metrics []*recorder.Metric) string {
	var buf bytes.Buffer

	if err := r.summaryTpl.Execute(&buf, map[string]interface{}{
		"Meta":      meta,
		"Customize": r.options,
		"I18n":      r.i18n,
		"Metrics":   metrics,
	}); err != nil {
		return fmt.Sprintf("%+v", errors.Wrap(err, "summaryTpl.Execute failed"))
	}

	return buf.String()
}

func (r *HtmlReporter) RenderUnit(meta *recorder.Meta, idx int, metric *recorder.Metric) string {
	var buf bytes.Buffer

	if err := r.unitTpl.Execute(&buf, map[string]interface{}{
		"Meta":      meta,
		"Customize": r.options,
		"I18n":      r.i18n,
		"Metric":    metric,
		"Idx":       idx,
	}); err != nil {
		return fmt.Sprintf("%+v", errors.Wrap(err, "unitTpl.Execute failed"))
	}

	return buf.String()
}

var reportTplStr = `<!DOCTYPE html>
<html lang="zh-cmn-Hans">
<head>
    <title>{{ .Test.Name }} {{ .I18n.Title.Report }}</title>
    <meta charset="UTF-8">
    <link href="https://cdn.jsdelivr.net/npm/bootstrap@5.1.3/dist/css/bootstrap.min.css" rel="stylesheet" integrity="sha384-1BmE4kWBq78iYhFldvKuhfTAU6auU8tT94WrHftjDbrCEXSU1oBoqyl2QvZ6jIW3" crossorigin="anonymous">
    <script src="https://cdn.jsdelivr.net/npm/bootstrap@5.1.3/dist/js/bootstrap.bundle.min.js" integrity="sha384-ka7Sk0Gln4gmtz2MlQnikT1wXgYsOg+OMhuP+IlRH9sENBO0LRn5q+8nbTov4+1p" crossorigin="anonymous"></script>
    <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/bootstrap-icons@1.8.1/font/bootstrap-icons.css">
    <script src="https://code.jquery.com/jquery-3.6.0.slim.min.js" integrity="sha256-u7e5khyithlIdTpu22PHhENmPcRdFiHRjhAuHcs05RI=" crossorigin="anonymous"></script>
    <script src="https://cdn.jsdelivr.net/npm/echarts@5.3.2/dist/echarts.min.js" integrity="sha256-7rldQObjnoCubPizkatB4UZ0sCQzu2ePgyGSUcVN70E=" crossorigin="anonymous"></script>

    {{ .Customize.Font.Style }}
    <style>
        body {
            font-family: {{ .Customize.Font.Body }};
        }
        pre, code {
            font-family: {{ .Customize.Font.Code }};
        }
    </style>

    <script>
    var yAxisLabelFormatter = {
        byte: (b) => {
          const units = ['B', 'KB', 'MB', 'GB', 'TB', 'PB', 'EB', 'ZB', 'YB'];
          let l = 0, n = parseInt(b, 10) || 0;
          while(n >= 1024 && ++l){
              n = n/1024;
          }
          return(n.toFixed(n < 10 && l > 0 ? 1 : 0) + ' ' + units[l]);
        },
        bit: (b) => {
          const units = ['b', 'Kb', 'Mb', 'Gb', 'Tb', 'Pb', 'Eb', 'Zb', 'Yb'];
          let l = 0, n = parseInt(b, 10) || 0;
          while(n >= 1024 && ++l){
              n = n/1024;
          }
          return(n.toFixed(n < 10 && l > 0 ? 1 : 0) + ' ' + units[l]);
        },
        percent: (v) => {
            return v + "%";
        },
        times: (v) => {
          const units = ['', 'K', 'M', 'G', 'T', 'P', 'E', 'Z', 'Y'];
          let l = 0, n = parseInt(v, 10) || 0;
          while(n >= 1024 && ++l){
              n = n/1024;
          }
          return(n.toFixed(n < 10 && l > 0 ? 1 : 0) + ' ' + units[l]);
        }
    }

    </script>

    {{ .Customize.Extra.Head }}
</head>

<body>
    {{ .Customize.Extra.BodyHeader }}

	{{/* summary */}}
    <div class="container">
        <div class="row justify-content-md-center">
			{{ RenderSummary $.Meta $.Metrics }}
        </div>

		<div class="row justify-content-md-center">
			{{ range $idx, $metric := $.Metrics }}
			{{ RenderUnit $.Meta $idx $metric }}
			{{ end }}
        </div>
    </div>

    {{ .Customize.Extra.BodyFooter }}
</body>
<script>
    var tooltipTriggerList = [].slice.call(document.querySelectorAll('[data-bs-toggle="tooltip"]'))
    var tooltipList = tooltipTriggerList.map(function (tooltipTriggerEl) {
      return new bootstrap.Tooltip(tooltipTriggerEl)
    })
</script>
</html>
`

var summaryTplStr = `
<div class="col-md-12" id="{{ .Meta.Name }}-summary">
	<table class="table table-striped">
		<thead>
			<tr class="text-center">
				<th>{{ .I18n.Title.Index }}</th>
				<th>{{ .I18n.Title.Unit }}</th>
				<th>{{ .I18n.Title.Total }}</th>
				<th>{{ .I18n.Title.QPS }}</th>
				<th>{{ .I18n.Title.AvgResTimeMs }}</th>
				<th>{{ .I18n.Title.SuccessRatePercent }}</th>
			</tr>
		</thead>
		<tbody>
			{{ range $idx, $metric := $.Metrics }}
			{{ range $key, $summary := $metric.Summary }}
			<tr class="text-center">
				<th>{{ $idx }}</th>
				<th>{{ $key }}({{ index (index $.Meta.Parallel $idx) $key }})</th>
				<td>{{ $summary.Total }}</td>
				<td>{{ $summary.QPS }}</td>
				<td>{{ FormatFloat $summary.AvgResTimeMs }}</td>
				<td>{{ FormatFloat $summary.SuccessRatePercent }}</td>
			</tr>
			{{ end }}
			{{ end }}
		</tbody>
	</table>
</div>
`

var unitTplStr = `
<div class="card border-success mx-0 px-0">
<div class="card-header justify-content-between d-flex"> No.{{ $.Idx }} </div>
<div class="col-md-12">
	<div class="card-body d-flex justify-content-center">
        <div class="col-md-12" id="{{ printf "%s-unit-%d-err-code-distribution" $.Meta.Name $.Idx }}" style="height: 300px;"></div>
        <script>
            echarts.init(document.getElementById("{{ printf "%s-unit-%d-err-code-distribution" $.Meta.Name $.Idx }}")).setOption({
              title: {
                text: "{{ .I18n.Title.ErrCodeDistribution }}",
                left: "center",
              },
              textStyle: {
                fontFamily: "{{ .Customize.Font.Echarts }}",
              },
              tooltip: {
                trigger: "item"
              },
              toolbox: {
                feature: {
                  saveAsImage: {
                    title: "{{ .I18n.Tooltip.Save }}"
                  }
                }
              },
              series: [
				{{ $idx := 0 }}
                {{ range $key, $errCodeDistribution := $.Metric.ErrCodeDistribution }}
                {
                  name: "{{ $key }}",
                  type: "pie",
                  radius: ['{{ EchartCodeRadius1 $idx (len $.Metric.ErrCodeDistribution) }}%', '{{ EchartCodeRadius2 $idx (len $.Metric.ErrCodeDistribution) }}%'],
                  avoidLabelOverlap: false,
                  label: {
                    show: false,
                    position: 'center'
                  },
                  emphasis: {
                    label: {
                      show: true,
                      fontSize: '20',
                      fontWeight: 'bold'
                    }
                  },
                  labelLine: {
                    show: false
                  },
                  data: {{ JsonMarshal (DictToItems $errCodeDistribution) }}
                },
				{{ $idx = plusOne $idx }}
                {{ end }}
              ]
            });
        </script>
    </div>
</div>

<div class="col-md-12">
	<div class="card-body d-flex justify-content-center">
        <div class="col-md-12" id="{{ printf "%s-unit-%d-qps" $.Meta.Name $.Idx }}" style="height: 300px;"></div>
        <script>
            echarts.init(document.getElementById("{{ printf "%s-unit-%d-qps" $.Meta.Name $.Idx }}")).setOption({
              title: {
                text: "{{ .I18n.Title.QPS }}",
                left: "center",
              },
              textStyle: {
                fontFamily: "{{ .Customize.Font.Echarts }}",
              },
              tooltip: {
                trigger: 'axis',
                show: true,
                axisPointer: {
                    type: "cross"
                }
              },
              toolbox: {
                feature: {
                  saveAsImage: {
                    title: "{{ .I18n.Tooltip.Save }}"
                  }
                }
              },
              xAxis: {
                type: "time",
              },
              yAxis: {
                type: "value",
              },
              series: [
                {{ range $key, $measurement := $.Metric.QPS }}
                {
                  name: "{{ $key }}",
                  type: "line",
                  smooth: true,
                  symbol: "none",
                  areaStyle: {},
                  data: {{ JsonMarshal (MeasurementToSerial $measurement) }}
                },
                {{ end }}
              ]
            });
        </script>
    </div>
</div>

<div class="col-md-12">
	<div class="card-body d-flex justify-content-center">
        <div class="col-md-12" id="{{ printf "%s-unit-%d-avg-res-time-ms" $.Meta.Name $.Idx }}" style="height: 300px;"></div>
        <script>
            echarts.init(document.getElementById("{{ printf "%s-unit-%d-avg-res-time-ms" $.Meta.Name $.Idx }}")).setOption({
              title: {
                text: "{{ .I18n.Title.AvgResTimeMs }}",
                left: "center",
              },
              textStyle: {
                fontFamily: "{{ .Customize.Font.Echarts }}",
              },
              tooltip: {
                trigger: 'axis',
                show: true,
                axisPointer: {
                    type: "cross"
                }
              },
              toolbox: {
                feature: {
                  saveAsImage: {
                    title: "{{ .I18n.Tooltip.Save }}"
                  }
                }
              },
              xAxis: {
                type: "time",
              },
              yAxis: {
                type: "value",
              },
              series: [
                {{ range $key, $measurement := $.Metric.AvgResTimeMs }}
                {
                  name: "{{ $key }}",
                  type: "line",
                  smooth: true,
                  symbol: "none",
                  areaStyle: {},
                  data: {{ JsonMarshal (MeasurementToSerial $measurement) }}
                },
                {{ end }}
              ]
            });
        </script>
    </div>
</div>

<div class="col-md-12">
	<div class="card-body d-flex justify-content-center">
        <div class="col-md-12" id="{{ printf "%s-unit-%d-success-rate-percent" $.Meta.Name $.Idx }}" style="height: 300px;"></div>
        <script>
            echarts.init(document.getElementById("{{ printf "%s-unit-%d-success-rate-percent" $.Meta.Name $.Idx }}")).setOption({
              title: {
                text: "{{ .I18n.Title.SuccessRatePercent }}",
                left: "center",
              },
              textStyle: {
                fontFamily: "{{ .Customize.Font.Echarts }}",
              },
              tooltip: {
                trigger: 'axis',
                show: true,
                axisPointer: {
                    type: "cross"
                }
              },
              toolbox: {
                feature: {
                  saveAsImage: {
                    title: "{{ .I18n.Tooltip.Save }}"
                  }
                }
              },
              xAxis: {
                type: "time",
              },
              yAxis: {
                type: "value",
              },
              series: [
                {{ range $key, $measurement := $.Metric.SuccessRatePercent }}
                {
                  name: "{{ $key }}",
                  type: "line",
                  smooth: true,
                  symbol: "none",
                  areaStyle: {},
                  data: {{ JsonMarshal (MeasurementToSerial $measurement) }}
                },
                {{ end }}
              ]
            });
        </script>
    </div>
</div>
</div>
`
