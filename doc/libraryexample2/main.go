/*
 * libraryexample2 uses change sets to modify in-situ instead of replacing the whole configuration.
 */
package main

import (
	"fmt"

	ipvsctl "github.com/aschmidt75/ipvsctl/integration"
)

func main() {

	// Create IPVSConfig struct
	cfg := ipvsctl.NewIPVSConfig()

	// Get the current table
	if err := cfg.Get(); err != nil {
		panic(err)
	}
	fmt.Printf("Current ipvs table: %+v\n", cfg)

	// set up apply options that limit our actions. We want to
	// add services and destinations, later delete services
	opts := ipvsctl.ApplyOpts{
		KeepWeights: true,
		AllowedActions: ipvsctl.ApplyActions{
			ipvsctl.ApplyActionAddService:     true,
			ipvsctl.ApplyActionAddDestination: true,
			ipvsctl.ApplyActionDeleteService:  true,
		},
	}

	// create a change set with a single change set item: add a service
	cs := ipvsctl.NewChangeSet()
	cs.AddChange(ipvsctl.ChangeSetItem{
		Type: ipvsctl.AddService,
		Service: &ipvsctl.Service{
			Address:   "tcp://127.0.0.1:9876",
			SchedName: "rr",
		},
	})
	// could cs.AddChange(...) here as well for other services etc.
	fmt.Printf("Change set: %+v\n", cs)

	// apply this change
	if err := cfg.ApplyChangeSet(cfg, cs, opts); err != nil {
		fmt.Println(err)
	}

	// get ipvs tables again, locate the service we just added
	// reason here is because for the modification of existing
	// things we need the real structures from ipvs, and not
	// just the string handles
	if err := cfg.Get(); err != nil {
		panic(err)
	}
	var myService *ipvsctl.Service
	for _, s := range cfg.Services {
		if s.Address == "tcp://127.0.0.1:9876" {
			myService = s
		}
	}
	if myService == nil {
		panic("unable to find my service?")
	}

	// add a destination (could have done that with the first change set as well)
	cs2 := ipvsctl.NewChangeSet()
	cs2.AddChange(ipvsctl.ChangeSetItem{
		Type:    ipvsctl.AddDestination,
		Service: myService,
		Destination: &ipvsctl.Destination{
			Address: "127.0.0.2:12340",
			Weight:  1000,
			Forward: "nat",
		},
	})
	fmt.Printf("Change set: %+v\n", cs)

	// apply this change as well
	if err := cfg.ApplyChangeSet(cfg, cs2, opts); err != nil {
		fmt.Println(err)
	}

	// get ipvs tables again and print out
	if err := cfg.Get(); err != nil {
		panic(err)
	}

	fmt.Printf("Current ipvs table:\n")
	for _, s := range cfg.Services {
		fmt.Printf("+ %s (%s)\n", s.Address, s.SchedName)
		for _, d := range s.Destinations {
			fmt.Printf(" -> %s (%s) w:%d\n", d.Address, d.Forward, d.Weight)
		}
	}

	// delete what we created
	cs3 := ipvsctl.NewChangeSet()
	cs3.AddChange(ipvsctl.ChangeSetItem{
		Type:    ipvsctl.DeleteService,
		Service: myService,
	})
	fmt.Printf("Change set: %+v\n", cs)

	// apply this change as well
	if err := cfg.ApplyChangeSet(cfg, cs3, opts); err != nil {
		fmt.Println(err)
	}

}
