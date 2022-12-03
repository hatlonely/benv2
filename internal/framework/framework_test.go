package framework

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/hatlonely/go-kit/config"
	"github.com/hatlonely/go-kit/refx"
	"github.com/hatlonely/go-kit/strx"
	. "github.com/smartystreets/goconvey/convey"
)

var testYaml = `
name: TestFramework
ctx:
  sh:
    type: Shell
    options: {}
source:
  src:
    type: Dict
    options:
      - key1: val1
        key2: val2
      - key1: val3
        key2: val4
plan:
  duration: 1s
  unit:
    - name: unit1
      parallel: 3
      step:
        - ctx: sh
          req:
            Command: echo -n ${KEY1} ${KEY2}
            Envs:
              "#KEY1": source.src.key1
              "#KEY2": source.src.key2
        - ctx: sh
          req:
            Command: echo -n ${KEY3} ${KEY4}
            Envs:
              "#KEY3": stat.Step[0].Res.Stdout
stat: ben.json
`

func TestFramework_RunPlan(t *testing.T) {
	Convey("TestFramework_RunPlan", t, func() {
		_ = ioutil.WriteFile("test.yaml", []byte(testYaml), 0755)
		defer os.RemoveAll("test.yaml")
		cfg, err := config.NewConfigWithSimpleFile("test.yaml", config.WithSimpleFileType("Yaml"))
		So(err, ShouldBeNil)
		var options Options
		So(cfg.Unmarshal(&options, refx.WithCamelName()), ShouldBeNil)
		fmt.Println(strx.JsonMarshalIndent(options))
		fw, err := NewFrameworkWithOptions(&options)
		So(err, ShouldBeNil)
		So(fw.Run(), ShouldBeNil)
	})
}
