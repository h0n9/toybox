package main

import (
	"fmt"
	"os"

	"github.com/h0n9/toybox/grayscalor"
)

func main() {
	args := os.Args
	if len(args) < 3 {
		fmt.Printf("%s <from-img> <to-img>", args[0])
		os.Exit(1)
	}
	fromFilename := args[1]
	toFilename := args[2]

	fromFile, err := os.Open(fromFilename)
	if err != nil {
		panic(err)
	}
	defer fromFile.Close()

	toFile, err := os.Create(toFilename)
	if err != nil {
		panic(err)
	}
	defer toFile.Close()

	err = grayscalor.Convert(fromFile, toFile, 30)
	if err != nil {
		panic(err)
	}
}
