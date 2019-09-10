package integration

import (
	"fmt"

	ipvs "github.com/aschmidt75/ipvsctl/ipvs"
	log "github.com/sirupsen/logrus"
)

// IPVSApplyError signal an error when applying a new configuration
type IPVSApplyError struct {
	what string
}

func (e *IPVSApplyError) Error() string {
	return fmt.Sprintf("Unable to apply new config: %s", e.what)
}

// Apply checks given ipvsconfig and applies it if possible
func (ipvsconfig *IPVSConfig) Apply(newconfig *IPVSConfig) error {

	ipvs, err := ipvs.New("")
	if err != nil {
		return &IPVSHandleError{}
	}
	log.Debugf("%#v\n", ipvs)
	defer ipvs.Close()

	// 1: iterate through all services in ipvsconfig. If
	// newconfig does not contain the service, remove it
	for _, service := range ipvsconfig.Services {
		found := false

		for _, newService := range newconfig.Services {
			if service.IsEqual(newService) {
				found = true
			}
		}

		if found == false {
			log.Debugf("Removing from current config: %s,%s\n", service.Address, service.SchedName)
			err = ipvs.DelService(service.service)
			if err != nil {
				log.Errorf("unable to delete service: %s", err)
				return &IPVSApplyError{what: "unable to delete service"}
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
			newIPVSService, err := newService.NewIpvsServiceStruct()
			if err != nil {
				log.Errorf("unable to add new service: %s", err)
				return &IPVSApplyError{what: "unable to add service"}
			}

			log.Debugf("newIPVSService=%#v\n", newIPVSService)

			err = ipvs.NewService(newIPVSService)
			if err != nil {
				log.Errorf("unable to add new service: %s", err)
				return &IPVSApplyError{what: "unable to add ipvs service"}
			}

			log.Debugf("added: %#v\n", newIPVSService)
		}
	}

	// 3: iterate through all services in ipvsconfig. If
	// newconfig does contain the service, compare it
	for _, service := range ipvsconfig.Services {
		for _, newService := range newconfig.Services {
			if service.IsEqual(newService) {
				log.Debugf("Comparing: %s,%s\n", service.Address, service.SchedName)

				// same scheduler?
				if service.SchedName != newService.SchedName {
					// no, update scheduler
					newIPVSService, err := newService.NewIpvsServiceStruct()
					if err != nil {
						log.Errorf("unable to edit new service: %s", err)
						return &IPVSApplyError{what: "unable to edit service"}
					}

					err = ipvs.UpdateService(newIPVSService)
					if err != nil {
						log.Errorf("unable to edit new service: %s", err)
						return &IPVSApplyError{what: "unable to edit ipvs service"}
					}

					log.Debugf("edited: %#v\n", newIPVSService)
				}
			}
		}
	}

	return nil
}
