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
			if service.IsEqual(newService) {
				found = true
			}
		}

		log.Debugf("Found=%d, activeService=%s in new config\n", found, service.Address)

		if found == false {
			res.AddChange(ChangeSetItem{
				Type: deleteService,
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
			if service.IsEqual(newService) {
				found = true
			}
		}

		if found == false {
			res.AddChange(ChangeSetItem{
				Type:        addService,
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
			if service.IsEqual(newService) {

				// same scheduler?
				if service.SchedName != newService.SchedName {
					// no, update service
					res.AddChange(ChangeSetItem{
						Type:        updateService,
						Service:     newService,
						Destination: nil,
					})
				}

				// iterate through destinations and compare

				// 1. walk through existing destinations. If newService does not contain it, remove.
				for _, destination := range service.Destinations {
					found := false
					for _, newDestination := range newService.Destinations {
						if destination.Address == newDestination.Address {
							found = true
						}
					}

					if found == false {
						res.AddChange(ChangeSetItem{
							Type: deleteDestination,
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
						if destination.Address == newDestination.Address {
							found = true
						}
					}
					if found == false {
						res.AddChange(ChangeSetItem{
							Type:        addDestination,
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
									Type:        updateDestination,
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
