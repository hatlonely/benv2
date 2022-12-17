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

//unit1: 3, unit2: 1
//
//|      summary       |     Total     |      QPS      | AvgResTimeMs | SuccessRatePercent |
//|--------------------|---------------|---------------|--------------|--------------------|
//|       unit1        |     1338      |    133.00     |    14.48     |       49.70        |
//|       unit2        |      444      |     45.00     |    14.52     |       50.68        |
//
//unit1: {"OK":668,"val3 val4":673}
//unit2: {"OK":225,"val3 val4":220}
//
//|        QPS         |  27:31  |  27:31  |  27:32  |  27:32  |  27:33  |  27:33  |  27:34  |  27:34  |  27:35  |  27:35  |  27:36  |
//|--------------------|---------|---------|---------|---------|---------|---------|---------|---------|---------|---------|---------|
//|       unit1        | 138.00  | 138.00  | 144.00  | 140.00  | 126.00  | 142.00  | 128.00  | 126.00  | 124.00  | 124.00  |  6.00   |
//|       unit2        |  52.00  |  46.00  |  52.00  |  44.00  |  42.00  |  48.00  |  42.00  |  42.00  |  38.00  |  44.00  |  0.00   |
//
//|    AvgResTimeMs    |  27:31  |  27:31  |  27:32  |  27:32  |  27:33  |  27:33  |  27:34  |  27:34  |  27:35  |  27:35  |  27:36  |
//|--------------------|---------|---------|---------|---------|---------|---------|---------|---------|---------|---------|---------|
//|       unit1        |  13.54  |  14.03  |  12.81  |  14.00  |  15.41  |  13.56  |  15.25  |  15.41  |  15.92  |  15.40  |  13.67  |
//|       unit2        |  12.85  |  13.87  |  13.31  |  13.05  |  16.90  |  13.54  |  14.95  |  15.95  |  15.68  |  15.95  |
//
//| SuccessRatePercent |  27:31  |  27:31  |  27:32  |  27:32  |  27:33  |  27:33  |  27:34  |  27:34  |  27:35  |  27:35  |  27:36  |
//|--------------------|---------|---------|---------|---------|---------|---------|---------|---------|---------|---------|---------|
//|       unit1        |  47.59  |  50.00  |  48.32  |  50.72  |  50.81  |  48.97  |  50.79  |  50.40  |  51.67  |  48.44  | 100.00  |
//|       unit2        |  56.52  |  50.00  |  54.17  |  47.83  |  48.84  |  52.17  |  47.73  |  50.00  |  44.19  |  55.00  |  0.00   |
//

var summaryTplStr = `
<div class="col-md-12" id="{{ .Meta.Name }}-summary">
	<table class="table table-striped">
		<thead>
			<tr class="text-center">
				<th>{{ .I18n.Title.Index }}</th>
				<th>{{ .I18n.Title.Unit }}</th>
				<th>{{ .I18n.Title.Parallel }}</th>
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
				<th>{{ $key }}</th>
				<th>{{ index (index $.Meta.Parallel $idx) $key }}</th>
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

<div class="col-md-12" id="{{ .Meta.Name }}-QPS">
errcode
</div>
`
