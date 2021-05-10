package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"syscall"

	"github.com/opencontainers/runtime-spec/specs-go"
	"github.com/pkg/errors"
	cli "github.com/urfave/cli/v2"

	"github.com/projecteru2/docker-cni/config"
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
	cli.VersionPrinter = func(_ *cli.Context) {
		printVersion()
	}

	app := &cli.App{
		Name: "docker-cni",
		Commands: []*cli.Command{
			{
				Name:  "oci",
				Usage: "run as oci wrapper",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "config",
						DefaultText: "/etc/docker/cni.yaml",
					},
				},
				Action: runOCI,
			},
			{
				Name:  "cni",
				Usage: "run as cni wrapper",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "cni",
						Usage: "cni binary filename",
					},
					&cli.StringFlag{
						Name:  "cni-config",
						Usage: "cni configure filename",
					},
					&cli.StringFlag{
						Name:        "logfile",
						Usage:       "record of cni stdout and stderr",
						DefaultText: "/var/log/cni.log",
					},
				},
				Action: runCNI,
			},
		},
	}
	if err := app.Run(os.Args); err != nil {
		fmt.Printf("Error running docker-cni: %+v\n", err)
		os.Exit(-1)
	}

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

func setup(configPath string, ociArgs []string) (conf config.Config, err error) {
	if conf, err = config.LoadConfig(configPath); err != nil {
		return
	}

	for i, args := range ociArgs {
		if args == "--bundle" {
			conf.OCISpecFilename = filepath.Join(ociArgs[i+1], "config.json")
		}
	}

	if err = conf.SetupLog(); err != nil {
		return
	}

	log.Debugf("config: %+v", conf)
	return conf, conf.Validate()
}

func runOCI(c *cli.Context) (err error) {
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
		if err = handleCreate(conf); err != nil {
			log.Errorf("failed to handle create: %+v", err)
			return
		}
	}

	return
}

func runCNI(c *cli.Context) error {
	stateBuf, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		return errors.Wrapf(err, "failed to read stdin: %+v", err)
	}
	var state specs.State
	if err = json.Unmarshal(stateBuf, &state); err != nil {
		return errors.Wrapf(err, "failed to unmarshal state json: %+v", err)
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
		return errors.Wrapf(err, "failed to open cni config %s: %+v", c.String("cni-config"), err)
	}
	if err := syscall.Dup2(int(file.Fd()), 0); err != nil {
		return errors.Wrapf(err, "failed to dup cni config to stdin: %+v", err)
	}

	if file, err = os.OpenFile(c.String("logfile"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err != nil {
		return errors.Wrapf(err, "failed to open logfile %s: %+v", c.String("logfile"), err)
	}
	if err := syscall.Dup2(int(file.Fd()), 1); err != nil {
		return errors.Wrapf(err, "failed to dup logfile to stdout: %v", err)
	}
	if err := syscall.Dup2(int(file.Fd()), 2); err != nil {
		return errors.Wrapf(err, "failed to dup logfile to stderr: %v", err)
	}

	return syscall.Exec(c.String("cni"), []string{c.String("cni")}, env)
}
