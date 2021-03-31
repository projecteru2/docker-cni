package oci

import (
	"github.com/opencontainers/runtime-spec/specs-go"
)

type ContainerMeta struct {
	bundlePath string
	specs.Spec
}

func LoadContainerMeta(bundlePath string) (*ContainerMeta, error) {
	return nil, nil
}
