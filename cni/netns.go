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

func (p netnsManager) AddNetns(ID string) (netnsPath string, add, del *utils.Process, err error) {
	if add, err = utils.NewProcess("ip", []string{"net", "a", p.getID(ID)}, nil, nil); err != nil {
		return
	}
	del, err = p.DelNetns(ID)
	return p.getNetnsPath(ID), add, del, err
}

func (p netnsManager) DelNetns(ID string) (*utils.Process, error) {
	return utils.NewProcess("ip", []string{"net", "d", p.getID(ID)}, nil, nil)
}

func (p netnsManager) getNetnsPath(ID string) string {
	return "/var/run/netns/" + p.getID(ID)
}

func (p netnsManager) getID(ID string) string {
	return ID[:12]
}
