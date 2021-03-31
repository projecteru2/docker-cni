package config

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type Config struct {
	CNIConfDir string `yaml:"cni_conf_dir" default:"/etc/cni/net.d/"`
	CNIBinDir  string `yaml:"cni_bin_dir" default:"/opt/cni/bin/"`
	LogDriver  string `yaml:"log_driver" default:"journal://"`
	LogLevel   string `yaml:"log_level" default:"info"`
}

func LoadConfig(path string) (config Config, err error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return
	}
	if err = yaml.Unmarshal(data, &config); err != nil {
		return
	}
	return
}
