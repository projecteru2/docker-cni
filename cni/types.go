package cni

import (
	"encoding/json"
	"io/ioutil"
	"os"
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
	Name string `json:"name"`
}

func NewCNIPlugin(specDir, binDir string) (_ *CNIPlugin, err error) {
	// walk thu the config_dir and get the first configure file in lexicographic order, the same behavior as kubelet: https://kubernetes.io/docs/concepts/extend-kubernetes/compute-storage-net/network-plugins/#cni

	filenames := []string{}
	if err = filepath.Walk(specDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		filenames = append(filenames, path)
		return nil
	}); err != nil {
		return nil, errors.WithStack(err)
	}

	if len(filenames) == 0 {
		return nil, errors.Errorf("cni configure not found: %s", specDir)
	}

	sort.Strings(filenames)
	content, err := ioutil.ReadFile(filenames[0])
	if err != nil {
		return nil, errors.WithStack(err)
	}

	spec := &CNISpec{}
	if err = json.Unmarshal(content, spec); err != nil {
		return nil, errors.WithStack(err)
	}

	return &CNIPlugin{
		binDir:    binDir,
		binPath:   filepath.Join(binDir, spec.Name),
		specBytes: content,
	}, nil
}

type netnsManager struct{}
