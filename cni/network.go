package cni

import (
	"github.com/pkg/errors"
	"github.com/projecteru2/docker-cni/config"
	"github.com/projecteru2/docker-cni/oci"
	"github.com/projecteru2/docker-cni/utils"
	log "github.com/sirupsen/logrus"
)

// checkupNetwork checks the netns and invokes "CNI check" (not implemented yet)
func (p *CNIPlugin) CheckupNetwork(conf config.Config, containerMeta *oci.ContainerMeta) (exist bool, cleanup []*utils.Process, err error) {
	netnsPath, err := p.GetNetns(containerMeta.ID)
	if err != nil {
		return false, nil, err
	}

	delCNI, err := p.DelCNI(containerMeta.ID, netnsPath, DefaultIfname)
	if err != nil {
		return false, nil, errors.WithStack(err)
	}
	delNetns, err := p.DelNetns(containerMeta.ID)
	if err != nil {
		return false, nil, errors.WithStack(err)
	}
	return true, []*utils.Process{delCNI, delNetns}, nil
}

func (p *CNIPlugin) SetupNetwork(conf config.Config, containerMeta *oci.ContainerMeta) (netnsPath string, cleanup []*utils.Process, err error) {
	netnsPath, addNetns, delNetns, err := p.AddNetns(containerMeta.ID)
	if err != nil {
		err = errors.WithStack(err)
		return
	}
	if _, _, err = addNetns.Run(); err != nil {
		err = errors.WithStack(err)
		return
	}
	defer func() {
		if err != nil {
			if _, _, e := delNetns.Run(); e != nil {
				log.Errorf("failed to rollback netns: %+v", e)
			}
		}
	}()

	addCNI, delCNI, err := p.AddCNI(containerMeta.ID, netnsPath, DefaultIfname)
	if err != nil {
		return
	}
	stdoutBytes, stderrBytes, err := addCNI.Run()
	log.Infof("add cni stdout: %s", string(stdoutBytes))
	log.Debugf("add cni stderr: %s", string(stderrBytes))
	return netnsPath, []*utils.Process{
		delCNI,
		delNetns,
	}, err
}
