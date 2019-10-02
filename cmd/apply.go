package cmd

import (
	"os"

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
			log.Errorf("Must specify an input file or - for stdin")
			os.Exit(exitInvalidFile)
		}

		// read new config from file
		newConfig, err := readModelFromInput(applyFile)
		if err != nil {
			os.Exit(exitValidateErr)
		}

		log.WithField("newconfig", newConfig).Debugf("read")

		// validate model before applying
		err = newConfig.Validate()
		if err != nil {
			log.Error(err)
			os.Exit(exitValidateErr)
		}

		// apply new configuration
		err = MustGetCurrentConfig().Apply(newConfig)
		if err != nil {
			log.Error(err)
			os.Exit(exitApplyErr)
		}
		log.Infof("Applied configuration from %s", *applyFile)
	}
}
