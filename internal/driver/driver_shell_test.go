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

		res, err := d.Do(map[string]interface{}{
			"Command": "echo -n hello world",
		})
		So(err, ShouldBeNil)
		So(res, ShouldResemble, map[string]interface{}{
			"Stdout":   "hello world",
			"Stderr":   "",
			"ExitCode": float64(0),
		})
	})
}
