package handler

import (
	"github.com/projecteru2/docker-cni/config"
	"github.com/projecteru2/docker-cni/oci"
)

type Handler interface {
	HandleCreate(config.Config, *oci.ContainerMeta) error
	HandleStart(config.Config, *oci.ContainerMeta) error
	HandleDelete(config.Config, *oci.ContainerMeta) error
	HandleCNIConfig([]byte) ([]byte, error)
}
