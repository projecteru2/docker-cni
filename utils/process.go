package utils

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"strings"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type Process struct {
	Path string
	Args []string
	Env  []string
	cmd  *exec.Cmd
}

func NewProcess(path string, args, env []string, stdin io.Reader) *Process {
	cmd := exec.Command(path, args...)
	if stdin == nil {
		stdin = os.Stdin
	}
	cmd.Stdin = stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = env
	return &Process{
		Path: path,
		Args: args,
		cmd:  cmd,
	}
}

func (p Process) Start() (err error) {
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
				log.Infof("forwarding signal: %s", sig.String())
				p.cmd.Process.Signal(sig)
			case <-ctx.Done():
				log.Debug("context cancelled, forwarding done")
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

func (p Process) Run() (err error) {
	if err = p.Start(); err != nil {
		return err
	}
	return p.Wait()
}

func (p Process) Command() string {
	return fmt.Sprintf("%s %s %s < %+v",
		strings.Join(p.Env, " "),
		p.Path,
		strings.Join(p.Args, " "),
		p.cmd.Stdin,
	)
}
