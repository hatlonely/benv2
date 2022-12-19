package main

import (
	"os"

	"github.com/hatlonely/go-kit/config"
	"github.com/hatlonely/go-kit/flag"
	"github.com/hatlonely/go-kit/refx"
	"github.com/hatlonely/go-kit/strx"

	"github.com/hatlonely/benv2/internal/framework"
)

var Version string

type Options struct {
	Help      bool   `flag:"-h; usage: show help info"`
	Version   bool   `flag:"-v; usage: show version"`
	Action    string `flag:"-a; default: run;usage: actions, one of [desc/run/analyst]"`
	Playbook  string `flag:"usage: playbook file; default: ben.yaml"`
	CamelName bool   `flag:"usage: use camel name as playbook field style"`
}

const (
	ECSuccess                = 0
	ECInvalidPlaybook        = 1
	ECUnmarshalOptionsFailed = 2
	ECFrameworkNewFailed     = 3
	ECFrameworkRunFailed     = 4
	ECFrameworkAnalystFailed = 5
)

func main() {
	var options Options
	refx.Must(flag.Struct(&options, refx.WithCamelName()))
	refx.Must(flag.Parse())
	if options.Help {
		strx.Trac(flag.Usage())
		strx.Trac(`
  ben -a run --playbook ben.yaml
`)
		return
	}
	if options.Version {
		strx.Trac(Version)
		return
	}

	cfg, err := config.NewConfigWithSimpleFile(options.Playbook, config.WithSimpleFileType("Yaml"))
	if err != nil {
		strx.Warn(err.Error())
		os.Exit(ECInvalidPlaybook)
	}

	var opts []refx.Option
	if options.CamelName {
		opts = append(opts, refx.WithCamelName())
	}

	var frameworkOptions framework.Options
	if err := cfg.Unmarshal(&frameworkOptions, opts...); err != nil {
		strx.Warn(err.Error())
		os.Exit(ECUnmarshalOptionsFailed)
	}

	if options.Action == "desc" {
		strx.Info(strx.JsonMarshalIndentSortKeys(frameworkOptions))
		os.Exit(ECSuccess)
	}

	fw, err := framework.NewFrameworkWithOptions(&frameworkOptions, opts...)
	if err != nil {
		strx.Warn(err.Error())
		os.Exit(ECFrameworkNewFailed)
	}

	if options.Action == "run" {
		if err := fw.Run(); err != nil {
			strx.Warn(err.Error())
			os.Exit(ECFrameworkRunFailed)
		}
	} else if options.Action == "analyst" {
		if err := fw.Analyst(); err != nil {
			strx.Warn(err.Error())
			os.Exit(ECFrameworkAnalystFailed)
		}
	}

	os.Exit(ECSuccess)
}
