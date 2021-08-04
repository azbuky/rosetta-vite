package main

import (
	"os"

	"github.com/azbuky/rosetta-vite/cmd"

	"github.com/fatih/color"
)

func main() {
	err := cmd.Execute()
	if err != nil {
		color.Red(err.Error())
		os.Exit(1)
	}
}
