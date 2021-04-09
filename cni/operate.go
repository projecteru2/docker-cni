package cni

import (
	"github.com/projecteru2/docker-cni/utils"
)

func (p *CNIPlugin) Add(ID, netnsPath, ifname string) (*utils.Process, error) {
	return utils.NewProcess(
		p.binPath, // path
		nil,       // args
		[]string{
			"CNI_COMMAND=ADD",
			"CNI_CONTAINERID=" + ID,
			"CNI_NETNS=" + netnsPath,
			"CNI_IFNAME=" + ifname,
			"CNI_PATH=" + p.binDir,
		}, // env
		utils.NewStdio(p.specBytes), // stdio
	)
}

func (p *CNIPlugin) Del(ID, netnsPath, ifname string) (*utils.Process, error) {
	return utils.NewProcess(
		p.binPath, // path
		nil,       // args
		[]string{
			"CNI_COMMAND=DEL",
			"CNI_CONTAINERID=" + ID,
			"CNI_NETNS=" + netnsPath,
			"CNI_IFNAME=" + ifname,
			"CNI_PATH=" + p.binDir,
		}, // env
		utils.NewStdio(p.specBytes), // stdio
	)
}
