package reporter

import (
	"html/template"

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

	funcs := template.FuncMap{
		"RenderTest": func() {},
	}

	reportTpl := template.Must(template.New("").Funcs(funcs).Parse(reportTplStr))

	return &HtmlReporter{
		i18n:      i18n_,
		options:   options,
		reportTpl: reportTpl,
	}, nil
}

type HtmlReporter struct {
	i18n    *i18n.I18n
	options *HtmlReporterOptions

	reportTpl *template.Template
}

func (r *HtmlReporter) Report(meta *recorder.Meta, metrics []*recorder.Metric) string {
	return "html"
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
    <div class="container">
        <div class="row justify-content-md-center">
            <div class="col-lg-10 col-md-12">
            {{ RenderTest .Test "test" }}
            </div>
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
