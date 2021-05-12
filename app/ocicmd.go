package app

import (
	"os"
	"path/filepath"
	"syscall"

	"github.com/projecteru2/docker-cni/config"
	"github.com/projecteru2/docker-cni/handler"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

func runOCI(handler handler.Handler) func(*cli.Context) error {
	return func(c *cli.Context) (err error) {
		configPath, ociArgs := c.String("config"), c.Args().Slice()

		conf, err := setup(configPath, ociArgs)
		if err != nil {
			log.Errorf("failed to setup: %+v", err)
			return
		}

		log.Infof("docker-cni running: %+v", os.Args)
		defer log.Infof("docker-cni finishing: %+v", err)

		switch parsePhase(ociArgs) {
		case CreatePhase:
			err = handler.HandleCreate(conf)

		case StartPhase:
			err = handler.HandleStart(conf)

		case DeletePhase:
			err = handler.HandleDelete(conf)
		}

		if err != nil {
			return
		}

		args := []string{conf.OCIBin}
		args = append(args, c.Args().Slice()...)
		syscall.Exec(conf.OCIBin, args, os.Environ())
		return
	}
}

func setup(configPath string, ociArgs []string) (conf config.Config, err error) {
	if conf, err = config.LoadConfig(configPath); err != nil {
		return
	}

	for i, args := range ociArgs {
		if args == "--bundle" {
			conf.OCISpecFilename = filepath.Join(ociArgs[i+1], "config.json")
		}
		if args == "--log" && conf.OCISpecFilename == "" {
			conf.OCISpecFilename = filepath.Join(filepath.Dir(ociArgs[i+1]), "config.json")
		}
	}

	if err = conf.SetupLog(); err != nil {
		return
	}

	log.Debugf("config: %+v", conf)
	return conf, conf.Validate()
}
