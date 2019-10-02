package integration

import (
	"fmt"

	ipvs "github.com/aschmidt75/ipvsctl/ipvs"
	log "github.com/sirupsen/logrus"
)

// IPVSApplyError signal an error when applying a new configuration
type IPVSApplyError struct {
	what    string
	origErr error
}

func (e *IPVSApplyError) Error() string {
	return fmt.Sprintf("Unable to apply new config: %s\nReason: %s", e.what, e.origErr)
}

// Apply checks given ipvsconfig and applies it if possible
func (ipvsconfig *IPVSConfig) Apply(newconfig *IPVSConfig) error {

	ipvs, err := ipvs.New("")
	if err != nil {
		return &IPVSHandleError{}
	}
	defer ipvs.Close()

	// create changeset from new configuration
	cs, err := ipvsconfig.ChangeSet(newconfig)
	if err != nil {
		return &IPVSApplyError{ what: "Unable to build change set from new configuration", origErr: err}
	}
	
	for idx, csiIntf := range cs.Items {
		csi := csiIntf.(ChangeSetItem)
		log.Debug("Applying change set item #%d (%#v)\n", idx, csi)

		switch csi.Type {
		case deleteService:
			log.Debugf("Removing service from current config: %s,%s\n", csi.Service.Address, csi.Service.SchedName)
			err = ipvs.DelService(csi.Service.service)
			if err != nil {
				return &IPVSApplyError{what: "unable to delete service", origErr: err}
			}
		case addService:
			log.Debugf("Adding to current config: %s,%s\n", csi.Service.Address, csi.Service.SchedName)

			newIPVSService, err := newconfig.NewIpvsServiceStruct(csi.Service)
			if err != nil {
				return &IPVSApplyError{what: "unable to add service", origErr: err}
			}

			log.Debugf("newIPVSService=%#v\n", newIPVSService)

			err = ipvs.NewService(newIPVSService)
			if err != nil {
				return &IPVSApplyError{what: "unable to add ipvs service", origErr: err}
			}

			log.Debugf("added: %#v\n", newIPVSService)

			newIPVSDestinations, err := newconfig.NewIpvsDestinationsStruct(csi.Service)
			if err != nil {
				return &IPVSApplyError{what: fmt.Sprintf("unable to add new destinations for service %s", csi.Service.Address), origErr: err}
			}

			for _, newIPVSDestination := range newIPVSDestinations {
				err = ipvs.NewDestination(newIPVSService, newIPVSDestination)
				if err != nil {
					return &IPVSApplyError{what: fmt.Sprintf("unable to add new destination %#v for service %s", newIPVSDestination.Address, csi.Service.Address), origErr: err}
				}
			}

		default:
			log.Debugf("Unhandled change type %s\n", csi.Type)			
		}
	}

	// 1: iterate through all services in ipvsconfig. If
	// newconfig does not contain the service, remove it
/*	for _, service := range ipvsconfig.Services {
		found := false

		for _, newService := range newconfig.Services {
			if service.IsEqual(newService) {
				found = true
			}
		}

		if found == false {
			log.Debugf("Removing service from current config: %s,%s\n", service.Address, service.SchedName)
			err = ipvs.DelService(service.service)
			if err != nil {
				return &IPVSApplyError{what: "unable to delete service", origErr: err}
			}
		}
	}

	// 2: iterate through all services in newconfig. If
	// current config does not contain the service, add it
	for _, newService := range newconfig.Services {
		found := false

		for _, service := range ipvsconfig.Services {
			if service.IsEqual(newService) {
				found = true
			}
		}

		if found == false {
			log.Debugf("Adding to current config: %s,%s\n", newService.Address, newService.SchedName)

			newIPVSService, err := newconfig.NewIpvsServiceStruct(newService)
			if err != nil {
				return &IPVSApplyError{what: "unable to add service", origErr: err}
			}

			log.Debugf("newIPVSService=%#v\n", newIPVSService)

			err = ipvs.NewService(newIPVSService)
			if err != nil {
				return &IPVSApplyError{what: "unable to add ipvs service", origErr: err}
			}

			log.Debugf("added: %#v\n", newIPVSService)

			newIPVSDestinations, err := newconfig.NewIpvsDestinationsStruct(newService)
			if err != nil {
				return &IPVSApplyError{what: fmt.Sprintf("unable to add new destinations for service %s", newService.Address), origErr: err}
			}

			for _, newIPVSDestination := range newIPVSDestinations {
				err = ipvs.NewDestination(newIPVSService, newIPVSDestination)
				if err != nil {
					return &IPVSApplyError{what: fmt.Sprintf("unable to add new destination %#v for service %s", newIPVSDestination.Address, newService.Address), origErr: err}
				}
			}
		}
	}
*/

	// 3: iterate through all services in ipvsconfig. If
	// newconfig does contain the service, compare it. If
	// anything differs, update it.
	for _, service := range ipvsconfig.Services {
		for _, newService := range newconfig.Services {
			if service.IsEqual(newService) {
				log.Debugf("Comparing: %s,%s\n", service.Address, service.SchedName)

				// same scheduler?
				if service.SchedName != newService.SchedName {
					// no, update scheduler
					newIPVSService, err := newconfig.NewIpvsServiceStruct(newService)
					if err != nil {
						return &IPVSApplyError{what: "unable to edit service", origErr: err}
					}

					err = ipvs.UpdateService(newIPVSService)
					if err != nil {
						return &IPVSApplyError{what: "unable to edit ipvs service", origErr: err}
					}
					log.Debugf("edited service: %#v\n", newIPVSService)
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
						log.Debugf("Removing destination from current config: %s\n", destination.Address)
						err = ipvs.DelDestination(service.service, destination.destination)
						if err != nil {
							return &IPVSApplyError{what: fmt.Sprintf("unable to delete destination %#s for service %s", destination.Address, service.Address), origErr: err}
						}
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
						log.Debugf("Adding to current config: %s to %s\n", newDestination.Address, service.Address)

						newIPVSDestination, err := newconfig.NewIpvsDestinationStruct(newDestination)
						if err != nil {
							return &IPVSApplyError{what: fmt.Sprintf("unable to prepare new destination for service %s", newService.Address), origErr: err}
						}
						err = ipvs.NewDestination(service.service, newIPVSDestination)
						if err != nil {
							return &IPVSApplyError{what: fmt.Sprintf("unable to add new destination %#v for service %s", newIPVSDestination.Address, newService.Address), origErr: err}
						}

					}
				}

				// 3. walk through existing destination. If newService contains it, compare both and edit if necessary.
				for _, destination := range service.Destinations {
					for _, newDestination := range newService.Destinations {
						if destination.Address == newDestination.Address {

							log.Debugf("Comparing: %#v\n", destination)

							//							log.Debugf("olddest=%#v\n", destination)
							//							log.Debugf("newdest=%#v\n", newDestination)

							if destination.Weight != newDestination.Weight ||
								destination.Forward != newDestination.Forward {
								log.Debugf("Updating: %s\n", destination.Address)

								updateIPVSDestination, err := newconfig.NewIpvsDestinationStruct(newDestination)
								if err != nil {
									return &IPVSApplyError{what: fmt.Sprintf("unable to prepare edited destination for service %s", newService.Address), origErr: err}
								}
								err = ipvs.UpdateDestination(service.service, updateIPVSDestination)
								if err != nil {
									return &IPVSApplyError{what: fmt.Sprintf("unable to update destination %#v for service %s", updateIPVSDestination.Address, newService.Address), origErr: err}
								}

							}
						}
					}
				}

			}
		}
	}

	return nil
}
