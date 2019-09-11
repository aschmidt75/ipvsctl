package cmd

import (
	"os"

	integration "github.com/aschmidt75/ipvsctl/integration"
	cli "github.com/jawher/mow.cli"
	log "github.com/sirupsen/logrus"
)

// Apply implements the "apply" cli command
func Apply(cmd *cli.Cmd) {
	cmd.Spec = "[-f=<FILENAME>]"
	var (
		applyFile = cmd.StringOpt("f", "/etc/ipvsctl.yaml", "File to apply. Use - for STDIN")
	)

	cmd.Action = func() {

		if *applyFile == "" {
			log.Errorf("Must specify an input file")
			os.Exit(exitInvalidFile)
		}
		log.Debugf("Using file=%s\n", *applyFile)

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
		newConfig, err := readModelFromInput(applyFile)
		if err != nil {
			os.Exit(exitValidateErr)
		}

		log.Debugf("newConfig=%#v\n", newConfig)

		// validate model before applying
		err = currentConfig.Validate()
		if err != nil {
			log.Error(err)
			os.Exit(exitValidateErr)
		}

		// apply new configuration
		err = currentConfig.Apply(newConfig)
		if err != nil {
			log.Error(err)
			os.Exit(exitApplyErr)
		}
	}
}
