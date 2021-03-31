package main

import (
	"os"

	"github.com/projecteru2/docker-cni/cni"
	"github.com/projecteru2/docker-cni/config"
	"github.com/projecteru2/docker-cni/oci"
	"github.com/projecteru2/docker-cni/utils"
	log "github.com/sirupsen/logrus"
)

func main() {
	if os.Args[1] == "--version" {
		printVersion()
		os.Exit(0)
	}

	configPath, bundlePath, ociPath, ociArgs, err := parseArgs(os.Args)
	if err != nil {
		log.Fatal("invalid arguments: %+v", err)
	}

	config, err := setup(configPath)
	if err != nil {
		log.Fatal("failed to setup: %+v", err)
	}

	containerMeta, err := oci.LoadContainerMeta(bundlePath)
	if err != nil {
		log.Fatal("failed to load container meta from oci spec: %+v", err)
	}

	netnsPath, err := setupNetwork(config, containerMeta)
	if err != nil {
		log.Fatal("failed to setup cni network: %+v", err)
	}

	if err := updateContainerMeta(config, containerMeta, netnsPath); err != nil {
		log.Fatal("failed to update oci spec: %+v", err)
	}

	os.Exit(runOCI(ociPath, ociArgs))
}

func printVersion() {}

func parseArgs(args []string) (configPath, bundlePath, ociPath string, ociArgs []string, err error) {
	return
}

func setup(configPath string) (conf config.Config, err error) {
	if conf, err = config.LoadConfig(configPath); err != nil {
		return
	}

	return conf, conf.SetupLog()
}

func setupNetwork(conf config.Config, containerMeta *oci.ContainerMeta) (netnsPath string, err error) {
	if netnsPath, err = cni.GetNetns(containerMeta.ID()); err == nil {
		log.Warnf("netns already exists, inherit %s", containerMeta.ID())
		return
	}

	if netnsPath, err = cni.CreateNetns(containerMeta.ID()); err != nil {
		return
	}
	return netnsPath, cni.Add(
		containerMeta.ID(),
		netnsPath,
		conf.CNIConfDir,
		conf.CNIBinDir,
		cni.AddOptions{containerMeta.Labels()},
	).Run()
}

func updateContainerMeta(conf config.Config, containerMeta *oci.ContainerMeta, netnsPath string) (err error) {
	containerMeta.UpdateNetns(netnsPath)
	containerMeta.AppendPoststopHook(*cni.Del(
		containerMeta.ID(),
		netnsPath,
		conf.CNIConfDir,
		conf.CNIBinDir,
		cni.DelOptions{},
	))
	containerMeta.Save()
	return nil
}

func runOCI(ociPath string, ociArgs []string) (returnCode int) {
	process := utils.NewProcess(ociPath, ociArgs, nil)
	return process.Run()
}
