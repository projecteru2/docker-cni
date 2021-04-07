package main

import (
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/projecteru2/docker-cni/cni"
	"github.com/projecteru2/docker-cni/config"
	"github.com/projecteru2/docker-cni/oci"
	"github.com/projecteru2/docker-cni/utils"
	log "github.com/sirupsen/logrus"
)

const (
	DefaultIfname = "eth0"
)

func main() {
	var err error

	// TODO@zc: --help / -h
	if len(os.Args) < 2 || os.Args[1] == "--version" {
		printVersion()
		os.Exit(0)
	}

	configPath, bundlePath, ociPath, ociArgs, err := parseArgs()
	if err != nil {
		log.Fatalf("invalid arguments: %+v", err)
	}

	conf, err := setup(configPath, bundlePath)
	if err != nil {
		log.Fatalf("failed to setup: %+v", err)
	}

	if conf.BundlePath != "" {
		// inject netns path and hook
		containerMeta, err := oci.LoadContainerMeta(conf.BundlePath)
		if err != nil {
			log.Fatalf("failed to load container meta from oci spec: %+v", err)
		}

		netnsPath, del, err := setupNetwork(conf, *containerMeta)
		if err != nil {
			log.Fatalf("failed to setup cni network: %+v", err)
		}
		// TODO@zc: this shall do in two oci processes
		defer func() {
			if err != nil {
				log.Info("rolling back, executing `%s`", del.Command())
				del.Run()
			}
		}()

		if err = updateContainerMeta(containerMeta, netnsPath, del); err != nil {
			log.Errorf("failed to update oci spec: %+v", err)
			return
		}
	}

	err = runOCI(ociPath, ociArgs)
	log.Debugf("docker-cni returns: %d", utils.ParseExitCode(err))
	os.Exit(utils.ParseExitCode(err))
}

func printVersion() {}

func parseArgs() (configPath, bundlePath, ociPath string, ociArgs []string, err error) {
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

		if arg == "--bundle" {
			bundlePath = os.Args[i+1]
		}
	}
	ociArgs = os.Args[idx:]

	if configPath == "" || ociPath == "" {
		err = errors.Errorf("--config, --runtime-path are required: %+v", os.Args)
	}
	return
}

func setup(configPath, bundlePath string) (conf config.Config, err error) {
	if conf, err = config.LoadConfig(configPath); err != nil {
		return
	}
	conf.BundlePath = filepath.Join(bundlePath, "config.json")
	return conf, conf.SetupLog()
}

func setupNetwork(conf config.Config, containerMeta oci.ContainerMeta) (netnsPath string, _ utils.Process, err error) {
	cniPlug, err := cni.NewCNIPlugin(conf.CNIConfDir, conf.CNIBinDir)
	if err != nil {
		return
	}

	if netnsPath, err = cniPlug.GetNetns(containerMeta.ID()); err == nil {
		log.Warnf("netns already exists, inherit %s", containerMeta.ID())
		// TODO@zc: may not be eth0
		return netnsPath, *cniPlug.Del(containerMeta.ID(), netnsPath, DefaultIfname), nil
	}

	if netnsPath, err = cniPlug.CreateNetns(containerMeta.ID()); err != nil {
		return
	}

	add, del := cniPlug.PairOperation(containerMeta.ID(), netnsPath, DefaultIfname)
	return netnsPath, *del, add.Run()
}

func updateContainerMeta(containerMeta *oci.ContainerMeta, netnsPath string, del utils.Process) (err error) {
	containerMeta.UpdateNetns(netnsPath)
	containerMeta.AppendPoststopHook(del)
	containerMeta.Save()
	return nil
}

func runOCI(ociPath string, ociArgs []string) (err error) {
	return utils.NewProcess(ociPath, ociArgs,
		nil, // env
		nil, // stdin
	).Run()
}
