package cni

import (
	"bytes"
	"io"

	"github.com/projecteru2/docker-cni/utils"
)

func (p *CNIPlugin) Add(ID, netnsPath, ifname string) *utils.Process {
	configReader := p.getConfigReader()
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
		configReader, // stdin
	)
}

func (p *CNIPlugin) Del(ID, netnsPath, ifname string) *utils.Process {
	return utils.NewProcess(
		p.binPath, // path
		nil,       // args
		[]string{
			"CNI_COMMAND=DEL",
			"CNI_CONTAINERID=" + ID,
			"CNI_IFNAME=" + ifname,
		}, // env
		nil, //stdin
	)
}

func (p *CNIPlugin) PairOperation(ID, netnsPath, ifname string) (addProcess, delProcess *utils.Process) {
	return p.Add(ID, netnsPath, ifname), p.Del(ID, netnsPath, ifname)
}

func (p *CNIPlugin) getConfigReader() io.Reader {
	return bytes.NewReader(p.specBytes)
}
