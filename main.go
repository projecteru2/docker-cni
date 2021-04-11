package main

import (
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/projecteru2/docker-cni/config"
	"github.com/projecteru2/docker-cni/utils"
	log "github.com/sirupsen/logrus"
)

type OCIPhase int

const (
	CreatePhase OCIPhase = iota
	StartPhase
	KillPhase
	DeletePhase
	OtherPhase
)

func main() {
	var err error
	defer func() {
		os.Exit(utils.ParseExitCode(err))
	}()

	// TODO@zc: --help / -h
	if len(os.Args) < 2 || os.Args[1] == "--version" || os.Args[1] == "-v" {
		printVersion()
		return
	}

	configPath, ociPath, ociArgs, err := parseArgs()
	if err != nil {
		log.Errorf("invalid arguments: %+v", err)
	}

	conf, err := setup(configPath, ociArgs)
	if err != nil {
		log.Errorf("failed to setup: %+v", err)
	}
	log.Debugf("docker-cni running: %+v", os.Args)

	//TODO: refine all log to "user story"

	var rollback func()
	switch parsePhase(ociArgs) {

	case CreatePhase:
		if rollback, err = handleCreate(conf); err != nil {
			log.Errorf("failed to handle create: %+v", err)
			return
		}

	case StartPhase:
		if rollback, err = handleStart(conf); err != nil {
			log.Errorf("failed to handle start: %+v", err)
			return
		}
	}
	if rollback != nil {
		defer func() {
			if err != nil {
				rollback()
			}
		}()
	}

	err = runOCI(ociPath, ociArgs)
	log.Debugf("docker-cni finishing: %+v", err)
}

func parsePhase(args []string) OCIPhase {
	for _, arg := range args {
		switch arg {
		case "create":
			return CreatePhase
		case "start":
			return StartPhase
		case "kill":
			return KillPhase
		case "delete":
			return DeletePhase
		}
	}
	return OtherPhase
}

func printVersion() {}

func parseArgs() (configPath, ociPath string, ociArgs []string, err error) {
	//TODO@zc: example config
	idx := 1
	for i, arg := range os.Args {
		if arg == "--config" {
			idx = i + 2
			configPath = os.Args[i+1]
			continue
		}

		if arg == "--runtime-path" {
			idx = i + 2
			ociPath = os.Args[i+1]
			continue
		}
	}
	ociArgs = os.Args[idx:]

	if configPath == "" || ociPath == "" {
		err = errors.Errorf("--config, --runtime-path are required: %+v", os.Args)
	}
	return
}

func setup(configPath string, ociArgs []string) (conf config.Config, err error) {
	if conf, err = config.LoadConfig(configPath); err != nil {
		return
	}

	conf.ID = ociArgs[len(ociArgs)-1]

	for i, args := range ociArgs {
		if args == "--bundle" {
			conf.OCISpecFilename = filepath.Join(ociArgs[i+1], "config.json")
			break
		}
		if args == "--log" {
			conf.OCISpecFilename = filepath.Join(filepath.Dir(ociArgs[i+1]), "config.json")
			break
		}
	}

	return conf, conf.SetupLog()
}
