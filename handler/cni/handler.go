package cni

import (
	"github.com/projecteru2/docker-cni/config"
	"github.com/projecteru2/docker-cni/oci"
)

type CNIHandler struct{}

func (a CNIHandler) HandleStart(_ config.Config, __ *oci.ContainerMeta) (err error) {
	return
}

func (a CNIHandler) HandleDelete(_ config.Config, __ *oci.ContainerMeta) (err error) {
	return
}
