package driver

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"github.com/hatlonely/go-kit/refx"
	"github.com/hatlonely/go-kit/strx"
	"github.com/pkg/errors"
)

type ShellDriverOptions struct {
	Shebang string            `dft:"bash"`
	Args    []string          `dft:"-c"`
	Envs    map[string]string
}

type ShellDriver struct{
	shebang string
	args    []string
	envs    []string
}

func NewShellDriverWithOptions(options *ShellDriverOptions) (*ShellDriver, error) {
	var envs []string
	for k, v := range options.Envs {
		envs = append(envs, fmt.Sprintf(`%s=%s`, k, strings.TrimSpace(v)))
	}

	return &ShellDriver{
		shebang: options.Shebang,
		args:    options.Args,
		envs:    envs,
	}, nil
}

type ShellDriverReq struct {
	Command string
	Envs    map[string]string
	Decoder string
}

type ShellDriverRes struct {
	Stdout   string
	Stderr   string
	ExitCode int
}

func (d *ShellDriver) Do(v interface{}) (interface{}, error) {
	req := &ShellDriverReq{}
	err := refx.InterfaceToStruct(v, &req)

	fmt.Println(strx.JsonMarshalIndentSortKeys(req))
	if err != nil {
		return nil, errors.WithMessage(err, "InterfaceToStruct failed")
	}
	return d.do(req)
}

func (d *ShellDriver) do(req *ShellDriverReq) (*ShellDriverRes, error) {
	var envs []string
	for k, v := range req.Envs {
		envs = append(envs, fmt.Sprintf(`%s=%s`, k, strings.TrimSpace(v)))
	}

	cmd := exec.Command(d.shebang, append(d.args, req.Command)...)
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, d.envs...)
	cmd.Env = append(cmd.Env, envs...)

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Start(); err != nil {
		return nil, errors.Wrap(err, "cmd.Start failed")
	}

	if err := cmd.Wait(); err != nil {
		exitCode := -1
		if e, ok := err.(*exec.ExitError); ok {
			if status, ok := e.Sys().(syscall.WaitStatus); ok {
				exitCode = status.ExitStatus()
			}
		}

		return &ShellDriverRes{
			Stdout:   stdout.String(),
			Stderr:   stderr.String(),
			ExitCode: exitCode,
		}, errors.Wrap(err, "cmd.Wait failed")
	}

	return &ShellDriverRes{
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		ExitCode: 0,
	}, nil
}