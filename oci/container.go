package oci

import (
	"encoding/json"
	"io/ioutil"

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
