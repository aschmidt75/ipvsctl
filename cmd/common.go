package cmd

import (
	"os"
	"io/ioutil"

	"gopkg.in/yaml.v2"

	integration "github.com/aschmidt75/ipvsctl/integration"
	log "github.com/sirupsen/logrus"
)

func readInput(filename *string) ([]byte, error) {
	var b []byte
	var err error
	if *filename == "-" {
		b, err = ioutil.ReadAll(os.Stdin)
		if err != nil {
			log.Errorf("Error reading from STDIN")
			os.Exit(exitInvalidFile)
		}
	} else {
		b, err = ioutil.ReadFile(*filename)
		if err != nil {
			log.Errorf("Error reading from input file %s", *filename)
			os.Exit(exitInvalidFile)
		}
	}

	return b, err
}

func readModelFromInput(filename *string) (*integration.IPVSConfig, error) {
	c := &integration.IPVSConfig{}

	b, err := readInput(filename)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(b, c)
	if err != nil {
		log.Errorf("Error parsing yaml")
		os.Exit(exitInvalidFile)
	}

	return c, err
}

// MustGetCurrentConfig queries the current IPVS configuration
// or exits in case of an error.
func MustGetCurrentConfig() *integration.IPVSConfig {
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
	return currentConfig
}
