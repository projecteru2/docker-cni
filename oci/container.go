package oci

import (
	"encoding/json"
	"io/ioutil"

	"github.com/opencontainers/runtime-spec/specs-go"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

func (c *ContainerMeta) Labels() map[string]string {
	return nil
}

func (c *ContainerMeta) UpdateNetns(netnsPath string) {
	for idx, ns := range c.Linux.Namespaces {
		if ns.Type == specs.NetworkNamespace {
			if ns.Path != "" {
				log.Warnf("netns path existed and have been replaced: %s", ns.Path)
			}
			c.Linux.Namespaces[idx] = specs.LinuxNamespace{
				Type: specs.NetworkNamespace,
				Path: netnsPath,
			}
		}
	}
}

func (c *ContainerMeta) AppendHook(phase, pathname string, args, env []string) {
	c.Hooks.Poststop = append(c.Hooks.Poststop, specs.Hook{
		Path: pathname,
		Args: args,
		Env:  env,
	})
}

func (c *ContainerMeta) Save() (err error) {
	data, err := json.Marshal(c.Spec)
	if err != nil {
		return errors.WithStack(err)
	}
	return errors.WithStack(ioutil.WriteFile(c.BundlePath, data, 0644))
}
