package driver

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"github.com/pkg/errors"
)

type ShellDriverOptions struct {
	Shebang string   `dft:"bash"`
	Args    []string `dft:"-c"`
	Envs    map[string]string
}

type ShellDriver struct {
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

type ShellDriverDoReq struct {
	Command    string
	Envs       map[string]string
	JsonDecode bool
}

type ShellDriverDoRes struct {
	Stdout   string
	Stderr   string
	ExitCode int
	Json     interface{}
}

func (d *ShellDriver) Do(req *ShellDriverDoReq) (*ShellDriverDoRes, error) {
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
		switch e := err.(type) {
		case *exec.ExitError:
			exitCode := -1
			if status, ok := e.Sys().(syscall.WaitStatus); ok {
				exitCode = status.ExitStatus()
			}
			return &ShellDriverDoRes{
				Stdout:   stdout.String(),
				Stderr:   stderr.String(),
				ExitCode: exitCode,
			}, nil
		}

		return nil, NewError(errors.Wrap(err, "cmd.Wait failed"), "CommandWaitFailed", err.Error())
	}

	if req.JsonDecode {
		var v interface{}
		if err := json.Unmarshal(stdout.Bytes(), &v); err != nil {
			return nil, NewError(errors.Wrap(err, "jsoniter.Unmarshal failed"), "JsonDecodeFailed", err.Error())
		}
		return &ShellDriverDoRes{
			Stderr:   stderr.String(),
			ExitCode: 0,
			Json:     v,
		}, nil
	}

	return &ShellDriverDoRes{
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		ExitCode: 0,
	}, nil
}
