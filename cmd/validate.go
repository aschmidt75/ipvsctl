package cmd

import (
	"fmt"
	"os"

	integration "github.com/aschmidt75/ipvsctl/integration"
	cli "github.com/jawher/mow.cli"
)

// Validate implements the "validate" cli command
func Validate(cmd *cli.Cmd) {
	cmd.Spec = "[-f=<FILENAME>]"
	var (
		filename = cmd.StringOpt("f", "/etc/ipvsctl.yaml", "File to apply. Use - for STDIN")
	)

	cmd.Action = func() {

		if *filename == "" {
			fmt.Fprintf(os.Stderr, "Must specify an input file\n")
			os.Exit(exitInvalidFile)
		}

		// read new config from file
		c, err := readModelFromInput(filename)
		if err != nil {
			os.Exit(exitValidateErr)
		}

		cr, err := resolveParams(c)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to resolve params: %s\n", err)
			os.Exit(exitParamErr)
		}

		//
		err = cr.Validate()
		if err != nil {
			e := err.(*integration.IPVSValidateError)
			fmt.Fprintf(os.Stderr, "Validation error: %s\n", e)
			os.Exit(exitValidateErr)
		}

		fmt.Println("Configuration valid.")
		os.Exit(exitOk)
	}
}
