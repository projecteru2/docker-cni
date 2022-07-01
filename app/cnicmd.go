package app

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"syscall"

	"github.com/opencontainers/runtime-spec/specs-go"
	"github.com/pkg/errors"
	"github.com/projecteru2/docker-cni/cni"
	"github.com/projecteru2/docker-cni/config"
	"github.com/projecteru2/docker-cni/handler"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

func runCNI(handler handler.Handler) func(*cli.Context) error {
	return func(c *cli.Context) (err error) {
		defer func() {
			if err != nil {
				log.Errorf("[hook] failed to preceed: %+v", err)
			}
		}()

		conf, err := config.LoadConfig(c.String("config"))
		if err != nil {
			return errors.WithStack(err)
		}

		if err = conf.SetupLog(); err != nil {
			return errors.WithStack(err)
		}

		stateBuf, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			return errors.WithStack(err)
		}
		var state specs.State
		if err = json.Unmarshal(stateBuf, &state); err != nil {
			return errors.WithStack(err)
		}

		file, err := os.OpenFile(conf.CNILog, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return errors.WithStack(err)
		}
		if err := syscall.Dup2(int(file.Fd()), 1); err != nil {
			return errors.WithStack(err)
		}
		if err := syscall.Dup2(int(file.Fd()), 2); err != nil {
			return errors.WithStack(err)
		}

		netns := ""
		if state.Pid != 0 {
			netns = fmt.Sprintf("/proc/%d/ns/net", state.Pid)
		}

		cniToolConfig := cni.CNIToolConfig{
			CNIPath:     conf.CNIBinDir,
			NetConfPath: conf.CNIConfDir,
			NetNS:       netns,
			Args:        os.Getenv("CNI_ARGS"),
			IfName:      conf.CNIIfname,
			Cmd:         c.String("command"),
			ContainerID: state.ID,
			Handler:     handler.HandleCNIConfig,
		}

		log.Infof("[hook] docker-cni running: %+v", cniToolConfig)
		err = cni.Run(cniToolConfig)
		return errors.WithStack(err)
	}
}
