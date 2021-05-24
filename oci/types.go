package oci

import (
	"encoding/json"
	"io/ioutil"
	"path"
	"strconv"
	"strings"

	"github.com/opencontainers/runtime-spec/specs-go"
	"github.com/pkg/errors"
)

type ContainerMeta struct {
	ID         string
	InitPid    int
	BundlePath string
	specs.Spec
}

func LoadContainerMeta(bundlePath string) (containerMeta *ContainerMeta, err error) {
	data, err := ioutil.ReadFile(bundlePath)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	pidData, err := ioutil.ReadFile(path.Join(path.Dir(bundlePath), "init.pid"))
	if err != nil {
		pidData = []byte{'0'}
	}

	initPid, err := strconv.Atoi(string(pidData))
	if err != nil {
		return
	}

	pathParts := strings.Split(path.Dir(bundlePath), "/")
	containerMeta = &ContainerMeta{
		ID:         pathParts[len(pathParts)-1],
		InitPid:    initPid,
		BundlePath: bundlePath,
	}
	return containerMeta, json.Unmarshal(data, &containerMeta.Spec)
}
