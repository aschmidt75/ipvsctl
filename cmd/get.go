package cmd

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"

	cli "github.com/jawher/mow.cli"
	log "github.com/sirupsen/logrus"
)

// Get implements the "get" cli command
func Get(cmd *cli.Cmd) {
	cmd.Action = func() {
		b, err := yaml.Marshal(MustGetCurrentConfig())
		if err != nil {
			log.Error("unable to format as yaml")
			os.Exit(exitErrOutput)
		}
		fmt.Printf("%s", string(b))
	}
}
