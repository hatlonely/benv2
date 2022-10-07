package driver

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

)


func TestShellDriver(t *testing.T) {
	Convey("TestShellDriver", t, func() {
		d, err := NewShellDriverWithOptions(&ShellDriverOptions{
			Shebang: "bash",
			Args:    []string{"-c"},
			Envs: map[string]string{},
		})
		So(err, ShouldBeNil)

		res, err := d.Do(map[string]interface{}{
			"Command": "echo -n hello world",
		})
		So(err, ShouldBeNil)
		So(res, ShouldResemble, &ShellDriverRes{
			Stdout:   "hello world",
			Stderr:   "",
			ExitCode: 0,
		})
	})
}

