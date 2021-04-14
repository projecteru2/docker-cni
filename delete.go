package main

import (
	"io/ioutil"

	"github.com/pkg/errors"
	"github.com/projecteru2/docker-cni/config"
	log "github.com/sirupsen/logrus"
)

func postHandleDelete(conf config.Config) (err error) {
	if conf.OCILogFilename == "" {
		return errors.Errorf("empty oci log filename")
	}

	logContent, err := ioutil.ReadFile(conf.OCILogFilename)
	if err != nil {
		return errors.WithStack(err)
	}
	log.Warnf("oci log: %s", string(logContent))
	return
}
