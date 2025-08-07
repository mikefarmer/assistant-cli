package main

import (
	"github.com/mikefarmer/assistant-cli/cmd"
)

var version = "dev"

func main() {
	cmd.SetVersion(version)
	cmd.Execute()
}
