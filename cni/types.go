package cni

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"
	"sort"

	"github.com/pkg/errors"
)

type CNIPlugin struct {
	netnsManager

	binDir    string
	binPath   string
	specBytes []byte
}

type CNISpec struct {
	Type string `json:"type"`
}

func NewCNIPlugin(specDir, binDir string) (_ *CNIPlugin, err error) {
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
		binDir:    binDir,
		binPath:   filepath.Join(binDir, spec.Type),
		specBytes: content,
	}, nil
}

type netnsManager struct{}
