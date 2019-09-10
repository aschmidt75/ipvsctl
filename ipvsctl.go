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

	app := cli.App("ipvsctl", "...")

	app.Version("version", "0.0.1")

	app.Spec = "[-d] [-v]"

	debug := app.BoolOpt("d debug", c.Debug, "Show debug messages")
	verbose := app.BoolOpt("v verbose", c.Verbose, "Show information. Default: true. False equals to being quiet")

	app.Command("get", "retrieve ipvs configuration and returns as yaml", cmd.Get)
	app.Command("apply", "apply a new configuration from file or stdin", cmd.Apply)

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
