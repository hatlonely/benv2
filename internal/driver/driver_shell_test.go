package driver

import (
	"testing"

	"github.com/hatlonely/go-kit/refx"
	. "github.com/smartystreets/goconvey/convey"
)

func TestShellDriver(t *testing.T) {
	Convey("TestShellDriver", t, func() {
		d, err := NewDriverWithOptions(&refx.TypeOptions{
			Type: "Shell",
			Options: &ShellDriverOptions{
				Shebang: "bash",
				Args:    []string{"-c"},
				Envs:    map[string]string{},
			},
		})
		So(err, ShouldBeNil)

		Convey("normal", func() {
			res, err := d.Do(map[string]interface{}{
				"Command": "echo -n hello world",
			})
			So(err, ShouldBeNil)
			So(res, ShouldResemble, map[string]interface{}{
				"Stdout":   "hello world",
				"Stderr":   "",
				"ExitCode": int64(0),
				"Json":     nil,
			})
		})

		Convey("normal with envs", func() {
			res, err := d.Do(map[string]interface{}{
				"Command": "echo -n ${KEY1} ${KEY2}",
				"Envs": map[string]string{
					"KEY1": "hello",
					"KEY2": "world",
				},
			})
			So(err, ShouldBeNil)
			So(res, ShouldResemble, map[string]interface{}{
				"Stdout":   "hello world",
				"Stderr":   "",
				"ExitCode": int64(0),
				"Json":     nil,
			})
		})

		Convey("normal with json decode", func() {
			res, err := d.Do(map[string]interface{}{
				"Command":    `echo -n '{"key1": "val1", "key2": "val2"}'`,
				"JsonDecode": true,
			})
			So(err, ShouldBeNil)
			So(res, ShouldResemble, map[string]interface{}{
				"Stdout":   "",
				"Stderr":   "",
				"ExitCode": int64(0),
				"Json": map[string]interface{}{
					"key1": "val1",
					"key2": "val2",
				},
			})
		})

		Convey("command not found", func() {
			res, err := d.Do(map[string]interface{}{
				"Command": "abc",
			})
			So(err, ShouldBeNil)
			So(res, ShouldResemble, map[string]interface{}{
				"Stdout":   "",
				"Stderr":   "bash: abc: command not found\n",
				"ExitCode": int64(127),
				"Json":     nil,
			})
		})

		Convey("json unmarshal failed", func() {
			res, err := d.Do(map[string]interface{}{
				"Command":    "echo -n hello world",
				"JsonDecode": true,
			})
			So(res, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.(*Error).Code, ShouldEqual, "JsonDecodeFailed")
		})
	})
}
