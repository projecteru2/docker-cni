package cni

import (
	"github.com/projecteru2/docker-cni/config"
)

type CNIHandler struct{}

func (a CNIHandler) HandleStart(_ config.Config) (err error) {
	return
}

func (a CNIHandler) HandleDelete(_ config.Config) (err error) {
	return
}
