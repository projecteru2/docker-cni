package config

import (
	"io/ioutil"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

type Config struct {
	CNIConfDir string `yaml:"cni_conf_dir" default:"/etc/cni/net.d/"`
	CNIBinDir  string `yaml:"cni_bin_dir" default:"/opt/cni/bin/"`
	CNIWrapper string `yaml:"cni_wrapper"`
	LogDriver  string `yaml:"log_driver" default:"file:///var/run/log/docker-cni.log"`
	LogLevel   string `yaml:"log_level" default:"info"`

	// from command line args
	OCISpecFilename string
	OCILogFilename  string
	ID              string
}

func LoadConfig(path string) (config Config, err error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return config, errors.WithStack(err)
	}
	if err = yaml.Unmarshal(data, &config); err != nil {
		return config, errors.WithStack(err)
	}
	return
}

func (c Config) Validate() error {
	if c.OCISpecFilename == "" {
		return errors.Errorf("invalid config: oci spec filename is required")
	}
	if c.ID == "" {
		return errors.Errorf("invalid config: container ID is required")
	}
	return nil
}
