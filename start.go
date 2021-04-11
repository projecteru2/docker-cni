package main

import (
	"github.com/pkg/errors"
	"github.com/projecteru2/docker-cni/config"
	"github.com/projecteru2/docker-cni/oci"
	log "github.com/sirupsen/logrus"
)

// handleStart merely generates rollback
func handleStart(conf config.Config) (rollback func(), err error) {
	containerMeta, err := oci.LoadContainerMeta(conf.ID, conf.OCISpecFilename)
	if err != nil {
		return nil, errors.Wrap(err, "failed to load container meta from oci spec")
	}

	_, cleanup, err := setupNetwork(conf, *containerMeta)
	return func() {
		for _, clean := range cleanup {
			if _, _, e := clean.Run(); e != nil {
				log.Errorf("failed to rollback %s: %+v", clean.String(), e)
			}
		}
	}, err
}
