package cmd

import (
	"fmt"
	"os"

	integration "github.com/aschmidt75/ipvsctl/integration"
	cli "github.com/jawher/mow.cli"
	log "github.com/sirupsen/logrus"
)

// Validate implements the "validate" cli command
func Validate(cmd *cli.Cmd) {
	cmd.Spec = "[-f=<FILENAME>]"
	var (
		filename = cmd.StringOpt("f", "/etc/ipvsctl.yaml", "File to apply. Use - for STDIN")
	)

	cmd.Action = func() {

		log.Debugf("Using file=%s\n", *filename)
		if *filename == "" {
			log.Errorf("Must specify an input file")
			os.Exit(exitInvalidFile)
		}

		// read new config from file
		c, err := readModelFromInput(filename)
		if err != nil {
			os.Exit(exitValidateErr)
		}

		log.Debugf("validateConfig=%#v\n", c)

		//
		err = c.Validate()
		if err != nil {
			e := err.(*integration.IPVSValidateError)
			fmt.Printf("Configuration contains errors: %s\n", e.What)
			os.Exit(exitValidateErr)
		}

		fmt.Println("Configuration valid.")
		os.Exit(exitOk)
	}
}
