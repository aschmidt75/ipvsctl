package cmd

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"

	integration "github.com/aschmidt75/ipvsctl/integration"
	cli "github.com/jawher/mow.cli"
	log "github.com/sirupsen/logrus"
)

// Get implements the "get" cli command
func Get(cmd *cli.Cmd) {
	cmd.Action = func() {

		res := &integration.IPVSConfig{}
		err := res.Get()
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

		b, err := yaml.Marshal(res)
		if err != nil {
			log.Error("unable to format as yaml")
			os.Exit(exitErrOutput)
		}
		fmt.Printf("%s", string(b))
	}
}
