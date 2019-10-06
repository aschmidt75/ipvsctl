package cmd

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"

	cli "github.com/jawher/mow.cli"
	log "github.com/sirupsen/logrus"
)

// ChangeSet implements the "changeset" cli command
func ChangeSet(cmd *cli.Cmd) {
	cmd.Spec = "[-f=<FILENAME>]"
	var (
		csFile = cmd.StringOpt("f", "/etc/ipvsctl.yaml", "File to compare against current state. Use - for STDIN")
	)

	cmd.Action = func() {

		if *csFile == "" {
			log.Errorf("Must specify an input file or - for stdin")
			os.Exit(exitInvalidFile)
		}

		// read new config from file
		newConfig, err := readModelFromInput(csFile)
		if err != nil {
			os.Exit(exitValidateErr)
		}

		log.Debugf("newConfig=%#v\n", newConfig)

		// validate model before applying
		err = newConfig.Validate()
		if err != nil {
			log.Error(err)
			os.Exit(exitValidateErr)
		}

		// create changeset from new configuration
		cs, err := MustGetCurrentConfig().ChangeSet(newConfig)
		if err != nil {
			log.Error(err)
			os.Exit(exitApplyErr)
		}

		b, err := yaml.Marshal(cs)
		if err != nil {
			log.Error("unable to format as yaml")
			os.Exit(exitErrOutput)
		}
		fmt.Printf("%s", string(b))
	}
}
