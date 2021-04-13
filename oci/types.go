package oci

import (
	"encoding/json"
	"io/ioutil"

	"github.com/opencontainers/runtime-spec/specs-go"
	"github.com/pkg/errors"
)

type ContainerMeta struct {
	ID         string
	BundlePath string
	specs.Spec
}

func LoadContainerMeta(ID, bundlePath string) (*ContainerMeta, error) {
	data, err := ioutil.ReadFile(bundlePath)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	containerMeta := &ContainerMeta{
		ID:         ID,
		BundlePath: bundlePath,
	}
	return containerMeta, json.Unmarshal(data, &containerMeta.Spec)
}
