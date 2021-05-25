package cni

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"
	"sort"

	"github.com/pkg/errors"
)

func FindCNI(specDir, binDir string) (cniFilename, cniConfigFilename string, err error) {
	// walk thru the config_dir and get the first configure file in lexicographic order, the same behavior as kubelet: https://kubernetes.io/docs/concepts/extend-kubernetes/compute-storage-net/network-plugins/#cni

	files, err := ioutil.ReadDir(specDir)
	if err != nil {
		return "", "", errors.WithStack(err)
	}

	if len(files) == 0 {
		return "", "", errors.Errorf("cni configure not found: %s", specDir)
	}

	sort.Slice(files, func(i, j int) bool { return files[i].Name() < files[j].Name() })
	cniConfigFilename = files[0].Name()

	content, err := ioutil.ReadFile(filepath.Join(specDir, cniConfigFilename))
	if err != nil {
		return "", "", errors.WithStack(err)
	}

	spec := &CNISpec{}
	if err = json.Unmarshal(content, spec); err != nil {
		return "", "", errors.WithStack(err)
	}

	return filepath.Join(binDir, spec.Type), filepath.Join(specDir, cniConfigFilename), nil
}
