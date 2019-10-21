package cmd

import (
	"errors"
	"fmt"
	"os"
	"strings"

	cli "github.com/jawher/mow.cli"
	log "github.com/sirupsen/logrus"
	integration "github.com/aschmidt75/ipvsctl/integration"

)

func parseAllowedActions(actionSpec *string) (integration.ApplyActions, error) {
	all := integration.ApplyActions{
		integration.ApplyActionAddService: true,
		integration.ApplyActionUpdateService: true,
		integration.ApplyActionDeleteService: true,
		integration.ApplyActionAddDestination: true,
		integration.ApplyActionUpdateDestination: true,
		integration.ApplyActionDeleteDestination: true,
	}
	if actionSpec != nil {
		if *actionSpec == "*" {
			return all, nil
		}

		actions := strings.Split(*actionSpec, ",")
		res := make(integration.ApplyActions,len(actions))
		for _, action := range actions {
			_, ex := all[integration.ApplyActionType(action)]
			if ex == false {
				// no such action
				return integration.ApplyActions{}, errors.New(fmt.Sprintf("Invalid action: %s", action))
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
		applyFile  = cmd.StringOpt("f", "/etc/ipvsctl.yaml", "File to apply. Use - for STDIN")
		keepWeights = cmd.BoolOpt("keep-weights", false, "Leave weights as they are when updating destinations")
		actionSpec = cmd.StringOpt("allowed-actions", "*", `
Comma-separated list of allowed actions.
as=Add service, us=update service, ds=delete service,
ad=Add destination, ud=update destination, dd=delete destination.
Default * for all actions.
`)
	)

	cmd.Action = func() {

		if *applyFile == "" {
			log.Errorf("Must specify an input file or - for stdin")
			os.Exit(exitInvalidFile)
		}

		// read new config from file
		newConfig, err := readModelFromInput(applyFile)
		if err != nil {
			os.Exit(exitValidateErr)
		}

		log.WithField("newconfig", newConfig).Debugf("read")

		// validate model before applying
		err = newConfig.Validate()
		if err != nil {
			log.Error(err)
			os.Exit(exitValidateErr)
		}

		allowedSet, err := parseAllowedActions(actionSpec)
		if err != nil {
			log.Error(err)
			os.Exit(exitInvalidInput)
		}
		log.WithField("allowedActions", allowedSet).Trace("parsed")

		// apply new configuration
		err = MustGetCurrentConfig().Apply(newConfig, integration.ApplyOpts{
			KeepWeights: *keepWeights,
			AllowedActions: allowedSet,
		})
		if err != nil {
			log.Error(err)
			os.Exit(exitApplyErr)
		}
		log.Infof("Applied configuration from %s", *applyFile)
	}
}
