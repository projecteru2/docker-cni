package cni

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"
	"sort"

	"github.com/pkg/errors"
)

const (
	DefaultIfname = "eth0"
)

type CNIPlugin struct {
	netnsManager

	wrapper   string
	binDir    string
	binPath   string
	extraEnv  []string
	specBytes []byte
}

type CNISpec struct {
	Type string `json:"type"`
}

func NewCNIPlugin(cniWrapper, specDir, binDir string, extraEnv []string) (_ *CNIPlugin, err error) {
	// walk thu the config_dir and get the first configure file in lexicographic order, the same behavior as kubelet: https://kubernetes.io/docs/concepts/extend-kubernetes/compute-storage-net/network-plugins/#cni

	files, err := ioutil.ReadDir(specDir)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	if len(files) == 0 {
		return nil, errors.Errorf("cni configure not found: %s", specDir)
	}

	sort.Slice(files, func(i, j int) bool { return files[i].Name() < files[j].Name() })
	content, err := ioutil.ReadFile(filepath.Join(specDir, files[0].Name()))
	if err != nil {
		return nil, errors.WithStack(err)
	}

	spec := &CNISpec{}
	if err = json.Unmarshal(content, spec); err != nil {
		return nil, errors.WithStack(err)
	}

	return &CNIPlugin{
		wrapper:   cniWrapper,
		binDir:    binDir,
		binPath:   filepath.Join(binDir, spec.Type),
		extraEnv:  extraEnv,
		specBytes: content,
	}, nil
}

type netnsManager struct{}
