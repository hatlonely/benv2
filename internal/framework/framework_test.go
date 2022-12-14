package framework

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"sync"
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
  duration: 5s
  interval: 1s
  parallel:
    - unit1: 3
      unit2: 1
    - unit1: 5
      unit2: 2
  unit:
    - name: unit1
      step:
        - ctx: sh
          success: res.Stdout == "val1 val2"
          errCode: res.Stdout
          req:
            Command: echo -n ${KEY1} ${KEY2}
            Envs:
              "#KEY1": source.src.key1
              "#KEY2": source.src.key2
        - ctx: sh
          req:
            Command: echo -n ${KEY3}
            Envs:
              "#KEY3": stat.Step[0].Res.Stdout
    - name: unit2
      step:
        - ctx: sh
          success: res.Stdout == "val1 val2"
          errCode: res.Stdout
          req:
            Command: echo -n ${KEY1} ${KEY2}
            Envs:
              "#KEY1": source.src.key1
              "#KEY2": source.src.key2
        - ctx: sh
          req:
            Command: echo -n ${KEY3}
            Envs:
              "#KEY3": stat.Step[0].Res.Stdout
recorder:
  type: File
  options:
    filePath: ben.json
    metaPath: meta.json

analyst:
  type: File
  options:
    filePath: ben.json
    metaPath: meta.json

statistics:
  pointNumber: 10

monitors:
  - type: ACM
    options:
      accessKeyId: xx
      accessKeySecret: xx
      regionId: cn-shanghai
      metrics:
        - statistic: Average
          dimension: instanceId
          namespace: acs_ecs_dashboard
          metric: CPUUtilization
          period: 60s
          matchers:
            - label: instanceId
              value: i-uf6d23s4d6t1x6rpdmld
              operator: EQUALS

reporter:
  type: Json
  options:
    valueWidth: 9
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
		fw, err := NewFrameworkWithOptions(&options, refx.WithCamelName())
		So(err, ShouldBeNil)
		So(fw.Run(), ShouldBeNil)
	})
}

func BenchmarkFileWriter(b *testing.B) {
	str := `{"Name":"unit1","Step":[{"Req":{"Command":"echo -n ${KEY1} ${KEY2}","Envs":{"KEY2":"val4","KEY1":"val3"}},"Res":{"Stdout":"val3 val4","Stderr":"","ExitCode":0},"Err":null,"ErrCode":"","ResTime":3716244},{"Req":{"Command":"echo -n ${KEY3} ${KEY4}","Envs":{"KEY3":"val3 val4"}},"Res":{"Stdout":"val3 val4","Stderr":"","ExitCode":0},"Err":null,"ErrCode":"","ResTime":3833361}],"ErrCode":"","ResTime":7631754}`
	parallel := 20
	bufsize := 32768

	b.Run("write raw file", func(b *testing.B) {
		b.SetParallelism(parallel)
		fp, err := os.Create("benchmark1.json")
		if err != nil {
			panic(err)
		}
		defer os.RemoveAll("benchmark1.json")
		defer fp.Close()

		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_, _ = fp.WriteString(str + "\n")
			}
		})
	})

	b.Run("write mutex", func(b *testing.B) {
		b.SetParallelism(parallel)

		fp, err := os.Create("benchmark2.json")
		if err != nil {
			panic(err)
		}
		defer os.RemoveAll("benchmark2.json")
		defer fp.Close()
		writer := bufio.NewWriterSize(fp, bufsize)
		defer writer.Flush()

		var mutex sync.Mutex
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				mutex.Lock()
				_, _ = writer.WriteString(str + "\n")
				mutex.Unlock()
			}
		})
	})

	b.Run("write channel", func(b *testing.B) {
		b.SetParallelism(parallel)

		fp, err := os.Create("benchmark3.json")
		if err != nil {
			panic(err)
		}
		defer os.RemoveAll("benchmark3.json")
		defer fp.Close()
		writer := bufio.NewWriterSize(fp, bufsize)
		defer writer.Flush()

		channel := make(chan string, 20)
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			for str := range channel {
				_, _ = writer.WriteString(str + "\n")
			}
			wg.Done()
		}()

		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				channel <- str
			}
		})

		close(channel)
		wg.Wait()
	})
}
