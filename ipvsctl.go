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

	app.Spec = "[-d] [-v] [--trace]"

	trace := app.BoolOpt("trace", c.Trace, "Show trace messages")
	debug := app.BoolOpt("d debug", c.Debug, "Show debug messages")
	verbose := app.BoolOpt("v verbose", c.Verbose, "Show more information")

	app.Command("get", "retrieve ipvs configuration", cmd.Get)

	app.Before = func() {
		if trace != nil {
			c.Trace = *trace
		}
		if debug != nil {
			c.Debug = *debug
		}
		if verbose != nil {
			c.Verbose = *verbose
		}
		logging.InitLogging(c.Trace, c.Debug, c.Verbose)
	}
	app.Run(os.Args)
}
