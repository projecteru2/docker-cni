package app

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"syscall"

	"github.com/opencontainers/runtime-spec/specs-go"
	"github.com/pkg/errors"
	"github.com/projecteru2/docker-cni/cni"
	"github.com/projecteru2/docker-cni/config"
	"github.com/urfave/cli/v2"
)

func runCNI(c *cli.Context) (err error) {
	stateBuf, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		return errors.WithStack(err)
	}
	var state specs.State
	if err = json.Unmarshal(stateBuf, &state); err != nil {
		return errors.WithStack(err)
	}

	conf, err := config.LoadConfig(c.String("config"))
	if err != nil {
		return
	}

	cniFilename, cniConfigFilename, err := cni.FindCNI(conf.CNIConfDir, conf.CNIBinDir)
	if err != nil {
		return
	}

	env := []string{
		"CNI_IFNAME=" + conf.CNIIfname,
		"CNI_PATH=" + conf.CNIBinDir,
		"CNI_ARGS=" + os.Getenv("CNI_ARGS"),
		"CNI_COMMAND=" + strings.ToUpper(c.String("command")),
		"CNI_CONTAINERID=" + state.ID,
	}

	if state.Pid != 0 {
		env = append(env, "CNI_NETNS="+fmt.Sprintf("/proc/%d/ns/net", state.Pid))
	}

	file, err := os.Open(cniConfigFilename)
	if err != nil {
		return errors.WithStack(err)
	}
	if err := syscall.Dup2(int(file.Fd()), 0); err != nil {
		return errors.WithStack(err)
	}

	if file, err = os.OpenFile(conf.CNILog, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err != nil {
		return errors.WithStack(err)
	}
	if err := syscall.Dup2(int(file.Fd()), 1); err != nil {
		return errors.WithStack(err)
	}
	if err := syscall.Dup2(int(file.Fd()), 2); err != nil {
		return errors.WithStack(err)
	}

	return errors.WithStack(syscall.Exec(cniFilename, []string{cniFilename}, env))
}
