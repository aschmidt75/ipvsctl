// libraryexample1 applies a model as a whole on top of an existing
// ipvs tables model.
//
package main

import (
	"fmt"

	ipvsctl "github.com/aschmidt75/ipvsctl/integration"
	"gopkg.in/yaml.v2"
)

const targetModel string = `
services:
- address: tcp://127.0.0.1:9876
  sched: rr
  destinations:
  - address: 127.0.0.2:12340
    forward: nat
  - address: 127.0.0.3:12340
    forward: nat
`

func main() {

	// Unmarshal targetModel into IPVSConfig struct
	var newConfig ipvsctl.IPVSConfig
	err := yaml.Unmarshal([]byte(targetModel), &newConfig)
	if err != nil {
		panic(err)
	}

	// Create IPVSConfig struct
	currentConfig := ipvsctl.NewIPVSConfig()

	// Get the current table
	if err := currentConfig.Get(); err != nil {
		panic(err)
	}
	fmt.Printf("Current ipvs table: %+v\n", currentConfig)

	// set up apply options that limit our actions. We want to
	// add services and destinations
	opts := ipvsctl.ApplyOpts{
		KeepWeights: true,
		AllowedActions: ipvsctl.ApplyActions{
			ipvsctl.ApplyActionAddService:     true,
			ipvsctl.ApplyActionAddDestination: true,
		},
	}

	// Apply the new configuration. This will automatically build a change
	// set and apply only changes that are necessary. That means if this
	// is run for the 2nd time, nothing gets changed because it's already in.
	if err := currentConfig.Apply(&newConfig, opts); err != nil {
		panic(err)
	}

	if err := currentConfig.Get(); err != nil {
		panic(err)
	}
	fmt.Printf("Current ipvs table:\n")
	for _, s := range currentConfig.Services {
		fmt.Printf("+ %s (%s)\n", s.Address, s.SchedName)
		for _, d := range s.Destinations {
			fmt.Printf(" -> %s (%s) w:%d\n", d.Address, d.Forward, d.Weight)
		}
	}
}
