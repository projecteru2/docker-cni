package config

import (
	"io/ioutil"
	"os"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

type Config struct {
	OCIBin string `yaml:"oci_bin" default:"/usr/bin/runc"`

	CNIConfDir string `yaml:"cni_conf_dir" default:"/etc/cni/net.d/"`
	CNIBinDir  string `yaml:"cni_bin_dir" default:"/opt/cni/bin/"`
	CNIIfname  string `yaml:"cni_ifname" default:"eth0"`
	CNILog     string `yaml:"cni_log" default:"/var/log/cni.log"`

	LogDriver string `yaml:"log_driver" default:"file:///var/log/docker-cni.log"`
	LogLevel  string `yaml:"log_level" default:"info"`

	// from command line args
	Filename        string
	BinPathname     string
	OCISpecFilename string
}

func LoadConfig(path string) (conf Config, err error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return conf, errors.WithStack(err)
	}
	if err = yaml.Unmarshal(data, &conf); err != nil {
		return conf, errors.WithStack(err)
	}
	conf.BinPathname = os.Args[0]
	conf.Filename = path
	return
}

func (c Config) Validate() error {
	if c.OCISpecFilename == "" {
		return errors.Errorf("invalid config: oci spec filename is required")
	}
	return nil
}
