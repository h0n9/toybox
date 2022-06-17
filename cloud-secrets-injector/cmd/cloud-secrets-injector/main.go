package main

import (
	"fmt"
	"os"

	"github.com/h0n9/toybox/cloud-secrets-injector/cli"
)

func main() {
	err := cli.RootCmd.Execute()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	os.Exit(0)
}
