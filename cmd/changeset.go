package cmd

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"

	integration "github.com/aschmidt75/ipvsctl/integration"
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
			log.Errorf("Must specify an input file")
			os.Exit(exitInvalidFile)
		}
		log.Debugf("Using file=%s\n", *csFile)

		// retrieve current config
		currentConfig := &integration.IPVSConfig{}
		err := currentConfig.Get()
		if err != nil {
			log.Error(err)

			if _, ok := err.(*integration.IPVSHandleError); ok {
				os.Exit(exitIpvsErrHandle)
			}
			if _, ok := err.(*integration.IPVSQueryError); ok {
				os.Exit(exitIpvsErrQuery)
			}
			os.Exit(exitUnknown)
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
		cs, err := currentConfig.ChangeSet(newConfig)
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
