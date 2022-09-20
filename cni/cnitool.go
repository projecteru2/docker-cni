// see:https://github.com/containernetworking/cni/blob/main/cnitool/cnitool.go
package cni

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"sort"
	"strings"

	"github.com/containernetworking/cni/libcni"
)

// Protocol parameters are passed to the plugins via OS environment variables.
const (
	CmdAdd   = "add"
	CmdCheck = "check"
	CmdDel   = "del"
)

// CNIToolConfig .
type CNIToolConfig struct {
	CNIPath        string `json:"cni_path"`
	NetConfPath    string `json:"net_conf_path"`
	NetNS          string `json:"net_ns"`
	Args           string `json:"args"`
	CapabilityArgs string `json:"capability_args"`
	IfName         string `json:"if_name"`
	Cmd            string `json:"cmd"`
	ContainerID    string `json:"container_id"`
	Handler        func([]byte) ([]byte, error)
}

func parseArgs(args string) ([][2]string, error) {
	var result [][2]string

	pairs := strings.Split(args, ";")
	for _, pair := range pairs {
		kv := strings.Split(pair, "=")
		if len(kv) != 2 || kv[0] == "" || kv[1] == "" {
			return nil, fmt.Errorf("invalid CNI_ARGS pair %q", pair)
		}

		result = append(result, [2]string{kv[0], kv[1]})
	}

	return result, nil
}

// ConfFromFile .
func ConfFromFile(filename string, handler func([]byte) ([]byte, error)) (*libcni.NetworkConfig, error) {
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("error reading %s: %w", filename, err)
	}
	if handler != nil {
		bytes, err = handler(bytes)
		if err != nil {
			return nil, fmt.Errorf("error handling %s: %w", filename, err)
		}
	}
	return libcni.ConfFromBytes(bytes)
}

// ConfListFromFile .
func ConfListFromFile(filename string, handler func([]byte) ([]byte, error)) (*libcni.NetworkConfigList, error) {
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("error reading %s: %w", filename, err)
	}
	if handler != nil {
		bytes, err = handler(bytes)
		if err != nil {
			return nil, fmt.Errorf("error handling %s: %w", filename, err)
		}
	}
	return libcni.ConfListFromBytes(bytes)
}

// LoadConf .
func LoadConf(dir string, handler func([]byte) ([]byte, error)) (*libcni.NetworkConfig, error) {
	files, err := libcni.ConfFiles(dir, []string{".conf", ".json"})
	switch {
	case err != nil:
		return nil, err
	case len(files) == 0:
		return nil, libcni.NoConfigsFoundError{Dir: dir}
	}
	sort.Strings(files)

	for _, confFile := range files {
		conf, err := ConfFromFile(confFile, handler)
		if err != nil {
			return nil, err
		}
		return conf, nil
	}
	return nil, libcni.NotFoundError{Dir: dir, Name: ""}
}

// LoadConfList .
func LoadConfList(dir string, handler func([]byte) ([]byte, error)) (*libcni.NetworkConfigList, error) {
	files, err := libcni.ConfFiles(dir, []string{".conflist"})
	if err != nil {
		return nil, err
	}
	sort.Strings(files)

	for _, confFile := range files {
		conf, err := ConfListFromFile(confFile, handler)
		if err != nil {
			return nil, err
		}
		return conf, nil
	}

	// Try and load a network configuration file (instead of list)
	// then upconvert.
	singleConf, err := LoadConf(dir, handler)
	if err != nil {
		// A little extra logic so the error makes sense
		if _, ok := err.(libcni.NoConfigsFoundError); len(files) != 0 && ok {
			// Config lists found but no config files found
			return nil, libcni.NotFoundError{Dir: dir, Name: ""}
		}

		return nil, err
	}
	return libcni.ConfListFromConf(singleConf)
}

// Run .
func Run(config CNIToolConfig) error {
	netconf, err := LoadConfList(config.NetConfPath, config.Handler)
	if err != nil {
		return err
	}

	var cniArgs [][2]string
	if len(config.Args) > 0 {
		cniArgs, err = parseArgs(config.Args)
		if err != nil {
			return err
		}
	}

	var capabilityArgs map[string]interface{}
	if len(config.CapabilityArgs) > 0 {
		if err = json.Unmarshal([]byte(config.CapabilityArgs), &capabilityArgs); err != nil {
			return err
		}
	}

	cninet := libcni.NewCNIConfig(filepath.SplitList(config.CNIPath), nil)

	rt := &libcni.RuntimeConf{
		ContainerID:    config.ContainerID,
		NetNS:          config.NetNS,
		IfName:         config.IfName,
		Args:           cniArgs,
		CapabilityArgs: capabilityArgs,
	}

	switch config.Cmd {
	case CmdAdd:
		result, err := cninet.AddNetworkList(context.TODO(), netconf, rt)
		if result != nil {
			_ = result.Print()
		}
		return err
	case CmdCheck:
		return cninet.CheckNetworkList(context.TODO(), netconf, rt)
	case CmdDel:
		return cninet.DelNetworkList(context.TODO(), netconf, rt)
	default:
		return fmt.Errorf("unsupported command %v", config.Cmd)
	}
}
