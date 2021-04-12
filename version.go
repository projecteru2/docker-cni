package main

import (
	"fmt"
	"runtime"
)

var (
	// NAME is app name
	NAME = "docker-cni"
	// VERSION is app version
	VERSION = "unknown"
	// REVISION is app revision
	REVISION = "HEAD"
	// BUILTAT is app built info
	BUILTAT = "now"
)

func printVersion() {
	fmt.Printf("Version:        %s\n", VERSION)
	fmt.Printf("Git hash:       %s\n", REVISION)
	fmt.Printf("Built:          %s\n", BUILTAT)
	fmt.Printf("Golang version: %s\n", runtime.Version())
	fmt.Printf("OS/Arch:        %s/%s\n", runtime.GOOS, runtime.GOARCH)

}
