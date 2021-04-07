package cni

import (
	"os"

	"github.com/pkg/errors"
	"github.com/projecteru2/docker-cni/utils"
)

func (p netnsManager) GetNetns(ID string) (netnsPath string, err error) {
	_, err = os.Stat(p.getNetnsPath(ID))
	return p.getNetnsPath(ID), errors.WithStack(err)
}

func (p netnsManager) CreateNetns(ID string) (netnsPath string, err error) {
	return p.getNetnsPath(ID), errors.WithStack(utils.NewProcess("ip", []string{"net", "a", p.getID(ID)}, nil, nil).Run())
}

func (p netnsManager) DeleteNetns(ID string) (err error) {
	return errors.WithStack(utils.NewProcess("ip", []string{"net", "d", p.getID(ID)}, nil, nil).Run())
}

func (p netnsManager) getNetnsPath(ID string) string {
	return "/var/run/netns/" + p.getID(ID)
}

func (p netnsManager) getID(ID string) string {
	return ID[:12]
}
