package main

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"
	"sort"

	"github.com/pkg/errors"
	"github.com/projecteru2/docker-cni/config"
	"github.com/projecteru2/docker-cni/oci"
)

type CNISpec struct {
	Type string `json:"type"`
}

func handleCreate(conf config.Config) (err error) {
	containerMeta, err := oci.LoadContainerMeta(conf.OCISpecFilename)
	if err != nil {
		return err
	}

	cniFilename, cniConfigFilename, err := findCNI(conf.CNIConfDir, conf.CNIBinDir)
	if err != nil {
		return err
	}

	containerMeta.AppendHook("prestart",
		conf.SelfPathname,
		[]string{conf.SelfPathname, "cni", "--cni", cniFilename, "--cni-config", cniConfigFilename}, // args
		[]string{
			"CNI_IFNAME=eth0",
			"CNI_PATH=" + conf.CNIBinDir,
			"CNI_COMMAND=ADD",
		}, // env
	)

	containerMeta.AppendHook("poststop",
		conf.SelfPathname,
		[]string{conf.SelfPathname, "cni", "--cni", cniFilename, "--cni-config", cniConfigFilename}, // args
		[]string{
			"CNI_IFNAME=eth0",
			"CNI_PATH=" + conf.CNIBinDir,
			"CNI_COMMAND=DEL",
		}, // env
	)

	return containerMeta.Save()

}

func findCNI(specDir, binDir string) (cniFilename, cniConfigFilename string, err error) {
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

	return filepath.Join(binDir, spec.Type), cniConfigFilename, nil
}
