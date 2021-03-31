package cni

import (
	"github.com/pkg/errors"
	"github.com/vishvananda/netns"
)

func (p netnsManager) GetNetns(ID string) (netnsPath string, err error) {
	ns, err := netns.GetFromName(ID)
	return ns.String(), errors.WithStack(err)
}

func (p netnsManager) CreateNetns(ID string) (netnsPath string, err error) {
	ns, err := netns.NewNamed(ID)
	// TODO@zc
	return ns.String(), errors.WithStack(err)
}
