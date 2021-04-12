package main

func printHelp() {
	help := `
docker-cni is an oci wrapper aimming to call CNI instead of CNM.

Usage:

	docker-cni --config /path/to/config --runtime /path/to/runc $@
	`
	println(help)
}
