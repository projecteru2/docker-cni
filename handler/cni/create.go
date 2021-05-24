package cni

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

func (h CNIHandler) HandleCreate(conf config.Config, containerMeta *oci.ContainerMeta) (err error) {
	if err = h.AddCNIStartHook(conf, containerMeta); err != nil {
		return
	}
	if err = h.AddCNIStopHook(conf, containerMeta); err != nil {
		return
	}
	return containerMeta.Save()
}

func (h CNIHandler) AddCNIStartHook(conf config.Config, containerMeta *oci.ContainerMeta) (err error) {
	cniFilename, cniConfigFilename, err := h.findCNI(conf.CNIConfDir, conf.CNIBinDir)
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
	return
}

func (h CNIHandler) AddCNIStopHook(conf config.Config, containerMeta *oci.ContainerMeta) (err error) {
	cniFilename, cniConfigFilename, err := h.findCNI(conf.CNIConfDir, conf.CNIBinDir)
	if err != nil {
		return err
	}

	containerMeta.AppendHook("poststop",
		conf.SelfPathname,
		[]string{conf.SelfPathname, "cni", "--cni", cniFilename, "--cni-config", cniConfigFilename}, // args
		[]string{
			"CNI_IFNAME=eth0",
			"CNI_PATH=" + conf.CNIBinDir,
			"CNI_COMMAND=DEL",
		}, // env
	)
	return
}

func (h CNIHandler) findCNI(specDir, binDir string) (cniFilename, cniConfigFilename string, err error) {
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
