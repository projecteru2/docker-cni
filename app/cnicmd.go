package app

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"syscall"

	"github.com/opencontainers/runtime-spec/specs-go"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
)

func runCNI(c *cli.Context) error {
	stateBuf, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		return errors.WithStack(err)
	}
	var state specs.State
	if err = json.Unmarshal(stateBuf, &state); err != nil {
		return errors.WithStack(err)
	}

	env := []string{
		"CNI_IFNAME=" + os.Getenv("CNI_IFNAME"),
		"CNI_PATH=" + os.Getenv("CNI_PATH"),
		"CNI_ARGS=" + os.Getenv("CNI_ARGS"),
		"CNI_COMMAND=" + os.Getenv("CNI_COMMAND"),
		"CNI_CONTAINERID=" + state.ID,
	}

	if state.Pid != 0 {
		env = append(env, "CNI_NETNS="+fmt.Sprintf("/proc/%d/ns/net", state.Pid))
	}

	file, err := os.Open(c.String("cni-config"))
	if err != nil {
		return errors.WithStack(err)
	}
	if err := syscall.Dup2(int(file.Fd()), 0); err != nil {
		return errors.WithStack(err)
	}

	if file, err = os.OpenFile(c.String("log"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err != nil {
		return errors.WithStack(err)
	}
	if err := syscall.Dup2(int(file.Fd()), 1); err != nil {
		return errors.WithStack(err)
	}
	if err := syscall.Dup2(int(file.Fd()), 2); err != nil {
		return errors.WithStack(err)
	}

	return errors.WithStack(syscall.Exec(c.String("cni"), []string{c.String("cni")}, env))
}
