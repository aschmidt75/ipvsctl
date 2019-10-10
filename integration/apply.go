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
	if e.origErr == nil {
		return fmt.Sprintf("Unable to apply new config: %s", e.what)
	}
	return fmt.Sprintf("Unable to apply new config: %s\nReason: %s", e.what, e.origErr)
}

// Apply compares new config to current config, builds a changeset and
// applies the change set items within.
func (ipvsconfig *IPVSConfig) Apply(newconfig *IPVSConfig) error {

	// create changeset from new configuration
	cs, err := ipvsconfig.ChangeSet(newconfig)
	if err != nil {
		return &IPVSApplyError{ what: "Unable to build change set from new configuration", origErr: err}
	}

	return ipvsconfig.ApplyChangeSet(newconfig, cs)
}

// ApplyChangeSet takes a chhange set and applies all change items to
// the given IPVSConfig 
func (ipvsconfig *IPVSConfig) ApplyChangeSet(newconfig *IPVSConfig, cs *ChangeSet) error {

	ipvs, err := ipvs.New("")
	if err != nil {
		return &IPVSHandleError{}
	}
	defer ipvs.Close()

	for idx, csiIntf := range cs.Items {
		csi := csiIntf.(ChangeSetItem)
		log.WithFields(log.Fields{
			"idx": idx, 
			"csi": csi,
		}).Tracef("Applying change set item")

		switch csi.Type {
		case DeleteService:
			log.WithFields(log.Fields{
				"addr": csi.Service.Address, 
				"sched": csi.Service.SchedName,
			}).Tracef("Removing service from current config")

			err = ipvs.DelService(csi.Service.service)
			if err != nil {
				return &IPVSApplyError{what: "unable to delete service", origErr: err}
			}
		case AddService:
			log.WithFields(log.Fields{
				"addr": csi.Service.Address, 
				"sched": csi.Service.SchedName,
			}).Tracef("Adding to current config")

			newIPVSService, err := newconfig.NewIpvsServiceStruct(csi.Service)
			if err != nil {
				return &IPVSApplyError{what: "unable to add service", origErr: err}
			}

			log.Tracef("newIPVSService=%#v\n", newIPVSService)

			err = ipvs.NewService(newIPVSService)
			if err != nil {
				return &IPVSApplyError{what: "unable to add ipvs service", origErr: err}
			}

			log.WithFields(log.Fields{
				"ipvssvc": newIPVSService, 
			}).Tracef("added")

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
		case UpdateService:
			log.WithFields(log.Fields{
				"addr": csi.Service.Address, 
			}).Tracef("Updating service")

			newIPVSService, err := newconfig.NewIpvsServiceStruct(csi.Service)
			if err != nil {
				return &IPVSApplyError{what: "unable to edit service", origErr: err}
			}

			err = ipvs.UpdateService(newIPVSService)
			if err != nil {
				return &IPVSApplyError{what: "unable to edit ipvs service", origErr: err}
			}
			log.Tracef("edited service: %#v\n", newIPVSService)
			
		case AddDestination:
			log.WithFields(log.Fields{
				"dst-addr": csi.Destination.Address, 
				"dst": csi.Destination,
				"svc-addr": csi.Service.Address,
			}).Tracef("Adding destination to current config")

			newIPVSDestination, err := newconfig.NewIpvsDestinationStruct(csi.Destination)
			if err != nil {
				return &IPVSApplyError{what: fmt.Sprintf("unable to prepare new destination for service %s", csi.Service.Address), origErr: err}
			}
			err = ipvs.NewDestination(csi.Service.service, newIPVSDestination)
			if err != nil {
				return &IPVSApplyError{what: fmt.Sprintf("unable to add new destination %#v for service %s", newIPVSDestination.Address, csi.Service.Address), origErr: err}
			}

		case DeleteDestination:
			log.WithFields(log.Fields{
				"dst-addr": csi.Destination.Address, 
				"dst": csi.Destination,
				"svc-addr": csi.Service.Address,
			}).Tracef("Removing destination from current config")

			err = ipvs.DelDestination(csi.Service.service, csi.Destination.destination)
			if err != nil {
				return &IPVSApplyError{what: fmt.Sprintf("unable to delete destination %s for service %s", csi.Destination.Address, csi.Service.Address), origErr: err}
			}

		case UpdateDestination:
			log.WithFields(log.Fields{
				"dst-addr": csi.Destination.Address, 
				"dst": csi.Destination,
				"svc-addr": csi.Service.Address,
			}).Tracef("Updating destination")

			updateIPVSDestination, err := newconfig.NewIpvsDestinationStruct(csi.Destination)
			if err != nil {
				return &IPVSApplyError{what: fmt.Sprintf("unable to prepare edited destination for service %s", csi.Service.Address), origErr: err}
			}
			log.Tracef("Updating destination: %#v\n", updateIPVSDestination)
			err = ipvs.UpdateDestination(csi.Service.service, updateIPVSDestination)
			if err != nil {
				return &IPVSApplyError{what: fmt.Sprintf("unable to update destination %#v for service %s", updateIPVSDestination.Address, csi.Service.Address), origErr: err}
			}

		default:
			log.WithField("type", csi.Type).Tracef("Unhandled change type")			
		}
	}

	return nil
}
