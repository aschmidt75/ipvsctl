package integration

import (
	log "github.com/sirupsen/logrus"
)

// ChangeSet compares current active configuration against newconfig and
// creates a change set containing all differences
func (ipvsconfig *IPVSConfig) ChangeSet(newconfig *IPVSConfig) (*ChangeSet, error) {

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

		log.Tracef("Found=%t, activeService=%s in new config\n", found, service.Address)

		if found == false {
			res.AddChange(ChangeSetItem{
				Type: DeleteService,
				Service: &Service{
					Address: MakeAdressStringFromIpvsService(service.service),
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
						res.AddChange(ChangeSetItem{
							Type: DeleteDestination,
							Destination: &Destination{
								Address:     MakeAdressStringFromIpvsDestination(destination.destination),
								destination: destination.destination,
							},
							Service: &Service{
								Address: MakeAdressStringFromIpvsService(service.service),
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
						res.AddChange(ChangeSetItem{
							Type:        AddDestination,
							Destination: newDestination,
							Service: &Service{
								Address: MakeAdressStringFromIpvsService(service.service),
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
							equal, err := CompareDestinationsEquality(ipvsconfig, destination, newconfig, newDestination)
							if err != nil {
								return res, err
							}
							if equal == false {

								res.AddChange(ChangeSetItem{
									Type:        UpdateDestination,
									Destination: newDestination,
									Service: &Service{
										Address: MakeAdressStringFromIpvsService(service.service),
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
