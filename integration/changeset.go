package integration

import (
	"fmt"

	log "github.com/sirupsen/logrus"
)

// ChangeSet compares current active configuration against newconfig and
// creates a change set containing all differences
func (ipvsconfig *IPVSConfig) ChangeSet(newconfig *IPVSConfig, opts ApplyOpts) (*ChangeSet, error) {

	res := NewChangeSet()

	// 1: iterate through all services in ipvsconfig. If
	// newconfig does not contain the service, remove it
	for _, service := range ipvsconfig.Services {
		found := false

		for _, newService := range newconfig.Services {
			equal, err := CompareServicesIdentifyingEquality(ipvsconfig, service, newconfig, newService)
			if err != nil {
				return res, err
			}
			if equal == true {
				found = true
			}
		}

		log.WithFields(log.Fields{
			"ex":            found,
			"activeService": service.Address,
		}).Tracef("In new config")

		if found == false {
			adr := MakeAdressStringFromIpvsService(service.service)
			res.AddChange(ChangeSetItem{
				Type:        DeleteService,
				Description: fmt.Sprintf("Delete existing service %s because it does not exist in updated model any more", adr),
				Service: &Service{
					Address: adr,
					service: service.service,
				},
				Destination: nil,
			})
		}
	}

	// 2: iterate through all services in newconfig. If
	// current config does not contain the service, add it
	// and all its destinations
	for _, newService := range newconfig.Services {
		found := false

		for _, service := range ipvsconfig.Services {
			equal, err := CompareServicesIdentifyingEquality(ipvsconfig, service, newconfig, newService)
			if err != nil {
				return res, err
			}
			if equal == true {
				found = true
			}
		}

		if found == false {
			res.AddChange(ChangeSetItem{
				Type:        AddService,
				Description: fmt.Sprintf("Adding new service %s because it does not yet exist", newService.Address),
				Service:     newService,
				Destination: nil,
			})
		}
	}

	// 3: iterate through all services in ipvsconfig. If
	// newconfig does contain the service, compare it. If
	// anything differs, update it.
	for _, service := range ipvsconfig.Services {
		for _, newService := range newconfig.Services {
			equal, err := CompareServicesIdentifyingEquality(ipvsconfig, service, newconfig, newService)
			if err != nil {
				return res, err
			}
			if equal == true {
				newSched := newService.SchedName
				if newSched == "" {
					// default given?
					if newconfig.Defaults.SchedName != nil {
						newSched = *newconfig.Defaults.SchedName
					}
				}
				if newSched == "" {
					// still empty? rr is the primary default scheduler
					newSched = "rr"
				}

				// same scheduler?
				if service.SchedName != newSched {
					// no, update service
					res.AddChange(ChangeSetItem{
						Type:        UpdateService,
						Description: fmt.Sprintf("Updating existing service %s because details have changed", newService.Address),
						Service:     newService,
						Destination: nil,
					})
				}

				// iterate through destinations and compare

				// 1. walk through existing destinations. If newService does not contain it, remove.
				for _, destination := range service.Destinations {
					found := false
					for _, newDestination := range newService.Destinations {
						equal, err := CompareDestinationIdentifyingEquality(ipvsconfig, destination, newconfig, newDestination)
						if err != nil {
							return res, err
						}
						if equal == true {
							found = true
						}
					}

					if found == false {
						adrDestination := MakeAdressStringFromIpvsDestination(destination.destination)
						adrService := MakeAdressStringFromIpvsService(service.service)
						res.AddChange(ChangeSetItem{
							Type:        DeleteDestination,
							Description: fmt.Sprintf("Delete existing destination %s in service %s because it does not exist in updated model any more", adrDestination, adrService),
							Destination: &Destination{
								Address:     adrService,
								destination: destination.destination,
							},
							Service: &Service{
								Address: adrDestination,
								service: service.service,
							},
						})
					}
				}

				// 2. walk through new destinations. If current service does not have it, add it.
				for _, newDestination := range newService.Destinations {
					found := false
					for _, destination := range service.Destinations {
						equal, err := CompareDestinationIdentifyingEquality(ipvsconfig, destination, newconfig, newDestination)
						if err != nil {
							return res, err
						}
						if equal == true {
							found = true
						}
					}
					if found == false {
						adrService := MakeAdressStringFromIpvsService(service.service)
						res.AddChange(ChangeSetItem{
							Type:        AddDestination,
							Description: fmt.Sprintf("Adding new destination %s to service %s because it does not yet exist", newDestination.Address, adrService),
							Destination: newDestination,
							Service: &Service{
								Address: adrService,
								service: service.service,
							},
						})
					}
				}

				// 3. walk through existing destination. If newService contains it,
				// compare both and edit if necessary.
				for _, destination := range service.Destinations {
					for _, newDestination := range newService.Destinations {

						if destination.Address == newDestination.Address {
							equal, err := CompareDestinationsEquality(ipvsconfig, destination, newconfig, newDestination, opts)
							if err != nil {
								return res, err
							}
							if equal == false {
								adrService := MakeAdressStringFromIpvsService(service.service)

								if opts.KeepWeights {
									// newDestination might have a new weight, but we keep the old one
									newDestination.Weight = destination.Weight
								}

								res.AddChange(ChangeSetItem{
									Type:        UpdateDestination,
									Description: fmt.Sprintf("Updating existing destination %s in service %s because details have changed", newDestination.Address, adrService),
									Destination: newDestination,
									Service: &Service{
										Address: adrService,
										service: service.service,
									},
								})

							}
						}
					}
				}

			}
		}
	}

	return res, nil
}
