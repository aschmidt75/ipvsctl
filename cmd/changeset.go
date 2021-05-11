package cmd

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"

	integration "github.com/aschmidt75/ipvsctl/integration"
	cli "github.com/jawher/mow.cli"
)

// ChangeSet implements the "changeset" cli command
func ChangeSet(cmd *cli.Cmd) {
	cmd.Spec = "[-f=<FILENAME>]"
	var (
		csFile = cmd.StringOpt("f", "/etc/ipvsctl.yaml", "File to compare against current state. Use - for STDIN")
	)

	cmd.Action = func() {

		if *csFile == "" {
			fmt.Fprintf(os.Stderr, "Must specify an input file or - for stdin\n")
			os.Exit(exitInvalidFile)
		}

		// read new config from file
		newConfig, err := readModelFromInput(csFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading model: %s\n", err)
			os.Exit(exitValidateErr)
		}

		resolvedConfig, err := resolveParams(newConfig)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error resolving parameters: %s\n", err)
			os.Exit(exitParamErr)
		}

		// validate model before applying
		err = resolvedConfig.Validate()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error validation model: %s\n", err)
			os.Exit(exitValidateErr)
		}

		// create changeset from new configuration
		cs, err := MustGetCurrentConfig().ChangeSet(resolvedConfig, integration.ApplyOpts{})
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error building/applying changeset: %s\n", err)
			os.Exit(exitApplyErr)
		}

		b, err := yaml.Marshal(cs)
		if err != nil {
			fmt.Fprintf(os.Stderr, "unable to format as yaml")
			os.Exit(exitErrOutput)
		}
		fmt.Printf("%s", string(b))
	}
}
