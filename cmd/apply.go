package cmd

import (
	"errors"
	"fmt"
	"os"
	"strings"

	integration "github.com/aschmidt75/ipvsctl/integration"
	cli "github.com/jawher/mow.cli"
)

func parseAllowedActions(actionSpec *string) (integration.ApplyActions, error) {
	all := integration.ApplyActions{
		integration.ApplyActionAddService:        true,
		integration.ApplyActionUpdateService:     true,
		integration.ApplyActionDeleteService:     true,
		integration.ApplyActionAddDestination:    true,
		integration.ApplyActionUpdateDestination: true,
		integration.ApplyActionDeleteDestination: true,
	}
	if actionSpec != nil {
		if *actionSpec == "*" {
			return all, nil
		}

		actions := strings.Split(*actionSpec, ",")
		res := make(integration.ApplyActions, len(actions))
		for _, action := range actions {
			_, ex := all[integration.ApplyActionType(action)]
			if ex == false {
				// no such action
				return integration.ApplyActions{}, fmt.Errorf("Invalid action: %s", action)
			}
			res[integration.ApplyActionType(action)] = true
		}
		return res, nil
	}
	return integration.ApplyActions{}, errors.New("internal error, no actionSpec given")
}

// Apply implements the "apply" cli command
func Apply(cmd *cli.Cmd) {
	cmd.Spec = "[-f=<FILENAME>] [--keep-weights] [--allowed-actions=<ACTIONS_SPEC>]"
	var (
		applyFile   = cmd.StringOpt("f", "/etc/ipvsctl.yaml", "File to apply. Use - for STDIN")
		keepWeights = cmd.BoolOpt("keep-weights", false, "Leave weights as they are when updating destinations")
		actionSpec  = cmd.StringOpt("allowed-actions", "*", `
Comma-separated list of allowed actions.
as=Add service, us=update service, ds=delete service,
ad=Add destination, ud=update destination, dd=delete destination.
Default * for all actions.
`)
	)

	cmd.Action = func() {

		if *applyFile == "" {
			fmt.Fprintf(os.Stderr, "Must specify an input file or - for stdin\n")
			os.Exit(exitInvalidFile)
		}

		// read new config from file
		newConfig, err := readModelFromInput(applyFile)
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

		allowedSet, err := parseAllowedActions(actionSpec)
		if err != nil {
			fmt.Fprintf(os.Stderr, "unable to process allowed actions: %s\n", err)
			os.Exit(exitInvalidInput)
		}

		// apply new configuration
		err = MustGetCurrentConfig().Apply(resolvedConfig, integration.ApplyOpts{
			KeepWeights:    *keepWeights,
			AllowedActions: allowedSet,
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error applying updates: %s\n", err)
			os.Exit(exitApplyErr)
		}
		fmt.Printf("Applied configuration from %s\n", *applyFile)
	}
}
