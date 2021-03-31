package main

import (
	"os"

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

	if os.Args[1] == "--version" {
		printVersion()
		os.Exit(0)
	}

	configPath, bundlePath, ociPath, ociArgs, err := parseArgs(os.Args)
	if err != nil {
		log.Fatal("invalid arguments: %+v", err)
	}

	conf, err := setup(configPath, bundlePath)
	if err != nil {
		log.Fatal("failed to setup: %+v", err)
	}

	containerMeta, err := oci.LoadContainerMeta(conf.BundlePath)
	if err != nil {
		log.Fatal("failed to load container meta from oci spec: %+v", err)
	}

	netnsPath, del, err := setupNetwork(conf, *containerMeta)
	if err != nil {
		log.Fatal("failed to setup cni network: %+v", err)
	}
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

	if err = runOCI(ociPath, ociArgs); err != nil {
		log.Errorf("failed to complete oci: %+v", err)
		return
	}
	os.Exit(utils.ParseExitCode(err))
}

func printVersion() {}

func parseArgs(args []string) (configPath, bundlePath, ociPath string, ociArgs []string, err error) {
	if args[1] == "--config" && len(args) > 2 {
		configPath = args[2]
	}
	if args[3] == "--runtime-path" && len(args) > 4 {
		ociPath = args[4]
	}
	ociArgs = args[5:]
	for idx, arg := range ociArgs {
		if arg == "--bundle" && len(ociArgs) > idx+1 {
			bundlePath = ociArgs[idx+1]
		}
	}
	if configPath == "" || bundlePath == "" || ociPath == "" {
		err = errors.Errorf("--config, --bundle, --runtime-path are required")
	}
	return
}

func setup(configPath, bundlePath string) (conf config.Config, err error) {
	if conf, err = config.LoadConfig(configPath); err != nil {
		return
	}
	conf.BundlePath = bundlePath
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
