package main

import (
	"fmt"
	"os"

	"github.com/projecteru2/docker-cni/app"
	"github.com/projecteru2/docker-cni/handler/cni"
)

func main() {
	app := app.NewApp(cni.CNIHandler{}, printVersion)
	if err := app.Run(os.Args); err != nil {
		fmt.Printf("Error running docker-cni: %+v\n", err)
		os.Exit(-1)
	}

}
