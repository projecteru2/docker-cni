package cni

import (
	"github.com/projecteru2/docker-cni/config"
	"github.com/projecteru2/docker-cni/oci"
)

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
	containerMeta.AppendHook("prestart",
		conf.BinPathname,
		[]string{conf.BinPathname, "cni", "--config", conf.Filename, "--command", "add"}, // args
		nil, // envs
	)
	return
}

func (h CNIHandler) AddCNIStopHook(conf config.Config, containerMeta *oci.ContainerMeta) (err error) {
	containerMeta.AppendHook("poststop",
		conf.BinPathname,
		[]string{conf.BinPathname, "cni", "--config", conf.Filename, "--command", "del"}, // args
		nil, // env
	)
	return
}
