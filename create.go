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

func handleCreate(conf config.Config) (rollback func(), err error) {
	containerMeta, err := oci.LoadContainerMeta(conf.ID, conf.OCISpecFilename)
	if err != nil {
		return nil, errors.Wrap(err, "failed to load container meta from oci spec")
	}

	netnsPath, cleanup, err := setupNetwork(conf, *containerMeta)
	if err != nil {
		return nil, errors.Wrap(err, "failed to setup cni network")
	}
	rollback = func() {
		for _, clean := range cleanup {
			if _, _, e := clean.Run(); e != nil {
				log.Errorf("failed to rollback %s: %+v", clean.String(), e)
			}
		}
	}
	defer func() {
		if err != nil {
			rollback()
		}
	}()
	return rollback, errors.Wrap(updateContainerMeta(containerMeta, netnsPath, cleanup), "failed to update oci spec")

}

// setupNetwork is meant to be idempotent
func setupNetwork(conf config.Config, containerMeta oci.ContainerMeta) (netnsPath string, cleanup []*utils.Process, err error) {
	cniPlug, err := cni.NewCNIPlugin(conf.CNIConfDir, conf.CNIBinDir)
	if err != nil {
		return
	}

	delCNI, err := cniPlug.Del(containerMeta.ID, netnsPath, DefaultIfname)
	if err != nil {
		return "", nil, errors.Wrap(err, "failed to init process to delete cni")
	}
	delNetns, err := cniPlug.DeleteNetns(containerMeta.ID)
	if err != nil {
		return "", nil, errors.Wrap(err, "failed to init process to delete netns")
	}
	cleanup = []*utils.Process{delCNI, delNetns}

	if netnsPath, err = cniPlug.GetNetns(containerMeta.ID); err == nil {
		log.Warnf("netns already exists, inherit %s", containerMeta.ID)
		return netnsPath, cleanup, nil
	}

	netnsPath, create, err := cniPlug.CreateNetns(containerMeta.ID)
	if err != nil {
		err = errors.Wrap(err, "failed to init process to create netns")
		return
	}
	if _, _, err = create.Run(); err != nil {
		err = errors.Wrap(err, "failed to run process to create netns")
		return
	}
	defer func() {
		if err != nil {
			if _, _, e := delNetns.Run(); e != nil {
				log.Errorf("failed to rollback netns: %+v", e)
			}
		}
	}()

	add, err := cniPlug.Add(containerMeta.ID, netnsPath, DefaultIfname)
	if err != nil {
		err = errors.Wrap(err, "failed to init process to add cni")
		return
	}
	stdoutBytes, stderrBytes, err := add.Run()
	log.Debugf("add cni stdout: %s", string(stdoutBytes))
	log.Debugf("add cni stderr: %s", string(stderrBytes))
	return netnsPath, cleanup, errors.Wrap(err, "failed to run process to add cni")
}

func updateContainerMeta(containerMeta *oci.ContainerMeta, netnsPath string, cleanup []*utils.Process) (err error) {
	containerMeta.UpdateNetns(netnsPath)
	for _, clean := range cleanup {
		containerMeta.AppendPoststopHook(clean)
	}
	containerMeta.Save()
	return nil
}

func runOCI(ociPath string, ociArgs []string) (err error) {
	stdio := utils.NewStdio(nil)
	proc, err := utils.NewProcess(ociPath, ociArgs,
		nil,   // env
		stdio, // stdio
	)
	if err != nil {
		return errors.Wrap(err, "failed to init process to run oci")
	}
	stdoutBytes, stderrBytes, err := proc.Run()
	os.Stdout.Write(stdoutBytes)
	os.Stderr.Write(stderrBytes)
	return err
}
