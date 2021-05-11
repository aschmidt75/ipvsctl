package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/aschmidt75/ipvsctl/integration"
	cli "github.com/jawher/mow.cli"
)

// Set implements the "set" cli command
func Set(cmd *cli.Cmd) {
	cmd.Command("weight", "set weight of a single destination", SetWeight)
}

// SetWeight implements the weight setting command
func SetWeight(cmd *cli.Cmd) {

	cmd.Spec = "WEIGHT --service=<SERVICE> --destination=<DESTINATION> [--time=<SECONDS>]"
	var (
		weight      = cmd.IntArg("WEIGHT", -1, "Weight [0..65535]")
		service     = cmd.StringOpt("s service", "", "Handle of service, e.g. tcp://127.0.0.1:80")
		destination = cmd.StringOpt("d destination", "", "Handle of destination, e.g. 10.0.0.1:80")
		timeSecs    = cmd.IntOpt("t time", 0, "Number of seconds, for drain/renew mode")
	)

	cmd.Action = func() {

		if *weight < 0 || *weight > 65535 {
			fmt.Fprintln(os.Stderr, "Invalid weight")
			os.Exit(exitInvalidInput)
		}

		if *service == "" {
			fmt.Fprintln(os.Stderr, "Service handle must not be empty")
			os.Exit(exitInvalidInput)
		}

		if *destination == "" {
			fmt.Fprintln(os.Stderr, "Destination handle must not be empty")
			os.Exit(exitInvalidInput)
		}

		if *timeSecs <= 0 {
			err := MustGetCurrentConfig().SetWeight(*service, *destination, *weight)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Unable to get current config: %s\n", err)
				os.Exit(exitSetErr)
			}
		} else {
			ch := make(integration.ContinousControlCh, 1)

			go func() {
				t := 0
				for t < *timeSecs {
					t = t + 1
					time.Sleep(1 * time.Second)
					ch <- integration.ControlAdvance
				}
				ch <- integration.ControlFinish
			}()

			err := MustGetCurrentConfig().SetWeightContinuous(*service, *destination, *weight, *timeSecs, ch)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Unable to set new weight: %s\n", err)
				os.Exit(exitSetErr)
			}
		}

	}
}
