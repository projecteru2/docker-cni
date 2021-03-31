package utils

import (
	"context"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type Process struct {
	Path string
	Args []string
	Env  []string
	cmd  *exec.Cmd
}

func NewProcess(path string, args []string, env []string) *Process {
	cmd := exec.Command(path, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = env
	return &Process{
		Path: path,
		Args: args,
		cmd:  cmd,
	}
}

func (p *Process) Start() (err error) {
	return errors.WithStack(p.cmd.Start())
}

func (p *Process) Wait() (returnCode int) {
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
	return parseReturnCode(errors.WithStack(p.cmd.Wait()))
}

func parseReturnCode(err error) int {
	if err == nil {
		return 0
	}
	if exit, ok := err.(*exec.ExitError); ok {
		if code, ok := exit.Sys().(syscall.WaitStatus); ok {
			return code.ExitStatus()
		}
	}
	log.Errorf("failed to retrieve exit code: %+v", err)
	return 1
}

func (p *Process) Run() (exitCode int) {
	if err := p.Start(); err != nil {
		log.Errorf("failed to start process: %+v", err)
		return 1
	}
	return p.Wait()
}
