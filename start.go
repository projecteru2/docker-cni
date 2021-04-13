package main

import (
	"github.com/pkg/errors"
	"github.com/projecteru2/docker-cni/cni"
	"github.com/projecteru2/docker-cni/config"
	"github.com/projecteru2/docker-cni/oci"
	log "github.com/sirupsen/logrus"
)

// handleStart merely generates rollback
func handleStart(conf config.Config) (rollback func(), err error) {
	containerMeta, err := oci.LoadContainerMeta(conf.ID, conf.OCISpecFilename)
	if err != nil {
		return
	}
	cniPlug, err := cni.NewCNIPlugin(conf.CNIConfDir, conf.CNIBinDir)
	if err != nil {
		return
	}
	exist, cleanup, err := cniPlug.CheckupNetwork(conf, containerMeta)
	if !exist {
		return nil, errors.Errorf("failed to checkup network: netns not exist")
	}
	return func() {
		for _, clean := range cleanup {
			if _, _, e := clean.Run(); e != nil {
				log.Errorf("failed to rollback %s: %+v", clean.String(), e)
			}
		}
	}, err
}
