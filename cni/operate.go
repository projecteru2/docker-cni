package cni

import (
	"strings"

	"github.com/projecteru2/docker-cni/utils"
)

func (p *CNIPlugin) addCNI(ID, netnsPath, ifname string) (*utils.Process, error) {
	return utils.NewProcess(
		p.binPath, // path
		nil,       // args
		[]string{
			"CNI_COMMAND=ADD",
			"CNI_CONTAINERID=" + ID,
			"CNI_NETNS=" + netnsPath,
			"CNI_IFNAME=" + ifname,
			"CNI_PATH=" + p.binDir,
			"CNI_ARGS=" + strings.Join(p.extraEnv, ";"),
		}, // env
		utils.NewStdio(p.specBytes), // stdio
	)
}

func (p *CNIPlugin) DelCNI(ID, netnsPath, ifname string) (*utils.Process, error) {
	path, args := p.wrapper, []string{p.binPath}
	if p.wrapper == "" {
		path, args = p.binPath, nil
	}
	return utils.NewProcess(
		path,
		args,
		[]string{
			"CNI_COMMAND=DEL",
			"CNI_CONTAINERID=" + ID,
			"CNI_NETNS=" + netnsPath,
			"CNI_IFNAME=" + ifname,
			"CNI_PATH=" + p.binDir,
			"CNI_ARGS=" + strings.Join(p.extraEnv, ";"),
		}, // env
		utils.NewStdio(p.specBytes), // stdio
	)
}

func (p *CNIPlugin) AddCNI(ID, netnsPath, ifname string) (add, del *utils.Process, err error) {
	if add, err = p.addCNI(ID, netnsPath, ifname); err != nil {
		return
	}
	del, err = p.DelCNI(ID, netnsPath, ifname)
	return
}
