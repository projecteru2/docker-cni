package config

import (
	"io/ioutil"
	"os"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

type Config struct {
	OCIBin     string `yaml:"oci_bin" default:"/usr/bin/runc"`
	CNIConfDir string `yaml:"cni_conf_dir" default:"/etc/cni/net.d/"`
	CNIBinDir  string `yaml:"cni_bin_dir" default:"/opt/cni/bin/"`
	LogDriver  string `yaml:"log_driver" default:"file:///var/run/log/docker-cni.log"`
	LogLevel   string `yaml:"log_level" default:"info"`

	// from command line args
	OCISpecFilename string

	SelfPathname string
}

func LoadConfig(path string) (conf Config, err error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return conf, errors.WithStack(err)
	}
	if err = yaml.Unmarshal(data, &conf); err != nil {
		return conf, errors.WithStack(err)
	}
	conf.SelfPathname = os.Args[0]
	return
}

func (c Config) Validate() error {
	if c.OCISpecFilename == "" {
		return errors.Errorf("invalid config: oci spec filename is required")
	}
	return nil
}
