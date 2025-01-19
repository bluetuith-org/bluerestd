package main

import (
	"os"

	"github.com/bluetuith-org/bluerestd/cmd/app"
)

func main() {
	app.New().Run(os.Args)
}
