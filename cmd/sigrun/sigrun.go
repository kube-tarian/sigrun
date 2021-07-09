package main

import (
	"log"

	cli "github.com/devopstoday11/sigrun/pkg/cli/commands"
)

func main() {
	err := cli.Run()
	if err != nil {
		log.Fatal(err)
	}
}
