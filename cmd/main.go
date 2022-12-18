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
	Help     bool   `flag:"-h; usage: show help info"`
	Version  bool   `flag:"-v; usage: show version"`
	Playbook string `flag:"usage: playbook file; default: ben.yaml"`
}

const (
	ECSuccess                = 0
	ECInvalidPlaybook        = 1
	ECUnmarshalOptionsFailed = 2
	ECFrameworkNewFailed     = 3
	ECFrameworkRunFailed     = 4
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

	var frameworkOptions framework.Options
	if err := cfg.Unmarshal(&frameworkOptions, refx.WithCamelName()); err != nil {
		strx.Warn(err.Error())
		os.Exit(ECUnmarshalOptionsFailed)
	}

	fw, err := framework.NewFrameworkWithOptions(&frameworkOptions, refx.WithCamelName())
	if err != nil {
		strx.Warn(err.Error())
		os.Exit(ECFrameworkNewFailed)
	}
	if err := fw.Run(); err != nil {
		strx.Warn(err.Error())
		os.Exit(ECFrameworkRunFailed)
	}

	os.Exit(ECSuccess)
}
