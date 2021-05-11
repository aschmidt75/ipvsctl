package cmd

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"

	cli "github.com/jawher/mow.cli"
)

// Get implements the "get" cli command
func Get(cmd *cli.Cmd) {
	cmd.Action = func() {
		b, err := yaml.Marshal(MustGetCurrentConfig())
		if err != nil {
			fmt.Fprintf(os.Stderr, "unable to format as yaml\n")
			os.Exit(exitErrOutput)
		}
		fmt.Printf("%s", string(b))
	}
}
