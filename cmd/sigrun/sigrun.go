package main

import (
	"log"

	"github.com/devopstoday11/sigrun/pkg/cli"
)

func main() {
	err := cli.Run()
	if err != nil {
		log.Fatal(err)
	}
}
