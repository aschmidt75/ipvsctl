package main

import (
	"os"

	"github.com/aschmidt75/ipvsctl/cmd"
	"github.com/aschmidt75/ipvsctl/config"
	"github.com/aschmidt75/ipvsctl/logging"
	cli "github.com/jawher/mow.cli"
)

func main() {
	c := config.Config()

	app := cli.App("ipvsctl", "A desired state configuration frontend for ipvs")

	app.Version("version", "0.1.0")

	app.Spec = "[-d] [-v]"

	debug := app.BoolOpt("d debug", c.Debug, "Show debug messages")
	verbose := app.BoolOpt("v verbose", c.Verbose, "Show information. Default: false. False equals to being quiet")

	app.Command("get", "retrieve ipvs configuration and returns as yaml", cmd.Get)
	app.Command("apply", "apply a new configuration from file or stdin", cmd.Apply)
	app.Command("validate", "validate a configuration from file or stdin", cmd.Validate)
	app.Command("changeset", "compare active ipvs configuration against file or stdin and return changeset", cmd.ChangeSet)
	app.Command("set", "change services and destinations", cmd.Set)

	app.Before = func() {
		if debug != nil {
			c.Debug = *debug
		}
		if verbose != nil {
			c.Verbose = *verbose
		}
		logging.InitLogging(false, c.Debug, c.Verbose)
	}
	app.Run(os.Args)
}
