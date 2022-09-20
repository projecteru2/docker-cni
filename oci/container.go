package oci

import (
	"encoding/json"
	"io/ioutil"
	"strings"

	"github.com/opencontainers/runtime-spec/specs-go"
	"github.com/pkg/errors"
)

func (c *ContainerMeta) AppendHook(phase, pathname string, args, env []string) {
	if c.Hooks == nil {
		c.Hooks = &specs.Hooks{}
	}
	newHook := specs.Hook{
		Path: pathname,
		Args: args,
		Env:  env,
	}
	switch phase {
	case "prestart":
		c.Hooks.Prestart = append(c.Hooks.Prestart, newHook)
	case "poststop":
		c.Hooks.Poststop = append(c.Hooks.Poststop, newHook)
	}
}

func (c *ContainerMeta) Save() (err error) {
	data, err := json.Marshal(c.Spec)
	if err != nil {
		return errors.WithStack(err)
	}
	return errors.WithStack(ioutil.WriteFile(c.BundlePath, data, 0644))
}

func (c *ContainerMeta) SpecificIP() string {
	for _, env := range c.Process.Env {
		parts := strings.Split(env, "=")
		if len(parts) == 2 && parts[0] == "IPV4" && parts[1] != "" {
			return parts[1]
		}
	}
	return ""
}

func (c *ContainerMeta) RequiresSpecificIP() bool {
	return c.SpecificIP() != ""
}

func (c *ContainerMeta) SpecificIPPool() string {
	for _, env := range c.Process.Env {
		parts := strings.Split(env, "=")
		if len(parts) == 2 && parts[0] == "IPPOOL" && parts[1] != "" {
			return parts[1]
		}
	}
	return ""
}

func (c *ContainerMeta) RequiresSpecificIPPool() bool {
	return c.SpecificIPPool() != ""
}

func (c *ContainerMeta) RequiresFixedIP() bool {
	for _, env := range c.Process.Env {
		parts := strings.Split(env, "=")
		if len(parts) == 2 && parts[0] == "FIXED_IP" && parts[1] != "0" {
			return true
		}
	}
	return false
}
