package utils

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"os/signal"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type Process struct {
	Path string
	Args []string
	Env  []string
	*Stdio

	cmd *exec.Cmd
}

type Stdio struct {
	StdinBytes []byte
	Stdout     io.Reader
	Stderr     io.Reader
}

func NewStdio(stdinBytes []byte) *Stdio {
	stdio := &Stdio{
		StdinBytes: stdinBytes,
	}
	return stdio
}

func (s *Stdio) Stdin() io.Reader {
	if s == nil || s.StdinBytes == nil {
		return os.Stdin
	}
	return bytes.NewReader(s.StdinBytes)
}

func (s *Stdio) StdoutBytes() []byte {
	bs, err := ioutil.ReadAll(s.Stdout)
	if err != nil {
		log.Errorf("failed to read stdout: %+v", err)
	}
	return bs
}

func (s *Stdio) StderrBytes() []byte {
	bs, err := ioutil.ReadAll(s.Stderr)
	if err != nil {
		log.Errorf("failed to read stderr: %+v", err)
	}
	return bs
}

func NewProcess(path string, args, env []string, stdio *Stdio) (_ *Process, err error) {
	cmd := exec.Command(path, args...)
	cmd.Stdin = stdio.Stdin()
	cmd.Env = env

	if stdio != nil {
		stdio.Stdout, err = cmd.StdoutPipe()
		if err != nil {
			return nil, errors.WithStack(err)
		}
		stdio.Stderr, err = cmd.StderrPipe()
		if err != nil {
			return nil, errors.WithStack(err)
		}
	}

	return &Process{
		Path:  path,
		Args:  args,
		Env:   env,
		Stdio: stdio,
		cmd:   cmd,
	}, nil
}

func (p Process) Start() (err error) {
	log.Debugf("spawning process %+v", p.cmd)
	return errors.WithStack(p.cmd.Start())
}

func (p Process) Wait() (err error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		signals := make(chan os.Signal, 32)
		signal.Notify(signals)
		for {
			select {
			case sig := <-signals:
				p.cmd.Process.Signal(sig)
			case <-ctx.Done():
				return
			}

		}
	}()
	return errors.WithStack(p.cmd.Wait())
}

func ParseExitCode(err error) int {
	if err == nil {
		return 0
	}
	if exit, ok := errors.Unwrap(err).(*exec.ExitError); ok {
		return exit.ExitCode()
	}
	log.Warnf("failed to retrieve exit code: %+v", err)
	return 1
}

func (p Process) Run() (stdoutBytes, stderrBytes []byte, err error) {
	if err = p.Start(); err != nil {
		return
	}
	if p.Stdio != nil {
		stdoutBytes, stderrBytes = p.Stdio.StdoutBytes(), p.Stdio.StderrBytes()
	}
	return stdoutBytes, stderrBytes, p.Wait()
}
