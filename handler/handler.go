package handler

import "github.com/projecteru2/docker-cni/config"

type Handler interface {
	HandleCreate(config.Config) error
	HandleStart(config.Config) error
	HandleDelete(config.Config) error
}
