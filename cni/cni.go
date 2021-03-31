package cni

import "github.com/projecteru2/docker-cni/utils"

func GetNetns(ID string) (netnsPath string, err error) {
	return
}

func CreateNetns(ID string) (netnsPath string, err error) {
	return
}

func Add(ID, netnsPath, confDir, binDir string, opts AddOptions) *utils.Process {
	return nil
}

func Del(ID, netnsPath, confDir, binDir string, opts DelOptions) *utils.Process {
	return nil
}
