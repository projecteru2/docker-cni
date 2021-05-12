package app

import (
	"github.com/projecteru2/docker-cni/handler"
	"github.com/urfave/cli/v2"
)

func NewApp(handler handler.Handler, printVersion func()) *cli.App {
	if printVersion != nil {
		cli.VersionPrinter = func(_ *cli.Context) {
			printVersion()
		}
	}

	return &cli.App{
		Name:    "docker-cni",
		Version: "-",
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
				Action: runOCI(handler),
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
						Name:  "log",
						Usage: "record of cni stdout and stderr",
						Value: "/var/log/cni.log",
					},
				},
				Action: runCNI,
			},
		},
	}
}
