package main

import (
	"os"

	"github.com/aschmidt75/ipvsctl/cmd"
	"github.com/aschmidt75/ipvsctl/config"
	cli "github.com/jawher/mow.cli"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	c := config.Config()

	app := cli.App("ipvsctl", "A desired state configuration frontend for ipvs")

	app.Version("version", version)

	app.Spec = "[-v] [--params-network] [--params-env] [--params-file=<FILE>...] [--params-url=<URL>...]"

	verbose := app.BoolOpt("v verbose", c.Verbose, "Show information. Default: false. False equals to being quiet")
	paramsHostNetwork := app.BoolOpt("params-network", c.ParamsHostNetwork, "Dynamic parameters. Add every network interface name as resolvable ip address, e.g. net.eth0")
	paramsHostEnv := app.BoolOpt("params-env", c.ParamsHostNetwork, "Dynamic parameters. Add every environment entry, e.g. env.port=<ENV VAR \"port\">")
	paramsFiles := make([]string, 10)
	app.StringsOptPtr(&paramsFiles, "params-file", []string{c.ParamsFilesFromEnv}, "Dynamic parameters. Add parameters from yaml or json file.")
	paramsURLs := make([]string, 10)
	app.StringsOptPtr(&paramsURLs, "params-url", []string{c.ParamsURLsFromEnv}, "Dynamic parameters. Add parameters from yaml or json resource given by URL.")

	app.Command("get", "retrieve ipvs configuration and returns as yaml", cmd.Get)
	app.Command("apply", "apply a new configuration from file or stdin", cmd.Apply)
	app.Command("validate", "validate a configuration from file or stdin", cmd.Validate)
	app.Command("changeset", "compare active ipvs configuration against file or stdin and return changeset", cmd.ChangeSet)
	app.Command("set", "change services and destinations", cmd.Set)

	app.Before = func() {
		if verbose != nil {
			c.Verbose = *verbose
		}
		c.SetupLogging()

		if paramsHostNetwork != nil {
			c.ParamsHostNetwork = *paramsHostNetwork
		}
		if paramsHostEnv != nil {
			c.ParamsHostEnv = *paramsHostEnv
		}
		c.ParamsFiles = make([]string, len(paramsFiles))
		copy(c.ParamsFiles, paramsFiles)
		c.ParamsURLs = make([]string, len(paramsURLs))
		copy(c.ParamsURLs, paramsURLs)
	}
	app.Run(os.Args)
}
