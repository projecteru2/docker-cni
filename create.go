package main

import (
	"os"

	"github.com/projecteru2/docker-cni/cni"
	"github.com/projecteru2/docker-cni/config"
	"github.com/projecteru2/docker-cni/oci"
	"github.com/projecteru2/docker-cni/utils"
	log "github.com/sirupsen/logrus"
)

func handleCreate(conf config.Config) (rollback func(), err error) {
	containerMeta, err := oci.LoadContainerMeta(conf.ID, conf.OCISpecFilename)
	if err != nil {
		return nil, err
	}
	cniPlug, err := cni.NewCNIPlugin(conf.CNIWrapper, conf.CNIConfDir, conf.CNIBinDir, containerMeta.Env())
	if err != nil {
		return
	}
	netnsPath, cleanup, err := cniPlug.SetupNetwork(conf, containerMeta)
	if err != nil {
		return nil, err
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
	return rollback, updateContainerMeta(containerMeta, netnsPath, cleanup)

}

func updateContainerMeta(containerMeta *oci.ContainerMeta, netnsPath string, cleanup []*utils.Process) (err error) {
	containerMeta.UpdateNetns(netnsPath)
	for _, clean := range cleanup {
		containerMeta.AppendPoststopHook(clean)
	}
	return containerMeta.Save()
}

func runOCI(ociPath string, ociArgs []string) (err error) {
	proc, err := utils.NewProcess(ociPath, ociArgs,
		nil,                 // env
		utils.NewStdio(nil), // stdio
	)
	if err != nil {
		return
	}
	stdoutBytes, stderrBytes, err := proc.Run()
	if _, e := os.Stdout.Write(stdoutBytes); e != nil {
		log.Errorf("failed to forward stdout from oci runtime: %+v", e)
	}
	if _, e := os.Stderr.Write(stderrBytes); e != nil {
		log.Errorf("failed to forward stderr from oci runtime: %+v", e)
	}
	return err
}
