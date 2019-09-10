package cmd

import (
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v2"

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

		log.Debugf("Using file=%s\n", *applyFile)
		if *applyFile == "" {
			log.Errorf("Must specify an input file")
			os.Exit(exitInvalidFile)
		}

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
		newConfig := &integration.IPVSConfig{}

		b, err := ioutil.ReadFile(*applyFile)
		if err != nil {
			log.Errorf("Error reading from input file %s", *applyFile)
			os.Exit(exitInvalidFile)
		}

		err = yaml.Unmarshal(b, newConfig)
		if err != nil {
			log.Errorf("Error parsing yaml from input file %s", *applyFile)
			os.Exit(exitInvalidFile)
		}

		log.Debugf("newConfig=%#v\n", newConfig)

		//
		err = currentConfig.Apply(newConfig)
		if err != nil {
			log.Error(err)
			os.Exit(exitApplyErr)
		}
	}
}
