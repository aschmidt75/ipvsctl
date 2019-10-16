package integration

import (
	"fmt"

	ipvs "github.com/aschmidt75/ipvsctl/ipvs"
	log "github.com/sirupsen/logrus"
)

// ApplyActionType is a mapped string to some action for the apply function
type ApplyActionType string

// ApplyActions maps actions to bools
type ApplyActions map[ApplyActionType]bool

const (
	// ApplyActionAddService allows the addition of new services
	ApplyActionAddService ApplyActionType = "as"

	// ApplyActionUpdateService allows the update of existing services
	ApplyActionUpdateService ApplyActionType = "us"

	// ApplyActionDeleteService allows deletion of existing services
	ApplyActionDeleteService ApplyActionType = "ds"

	// ApplyActionAddDestination allows the addition of new destinations to existing services
	ApplyActionAddDestination ApplyActionType = "ad"

	// ApplyActionUpdateDestination allows for updates of existing destinations
	ApplyActionUpdateDestination ApplyActionType = "ud"

	// ApplyActionDeleteDestination allows for deleting of existing destinations
	ApplyActionDeleteDestination ApplyActionType = "dd"
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
func (ipvsconfig *IPVSConfig) Apply(newconfig *IPVSConfig, allowedActions ApplyActions) error {

	// create changeset from new configuration
	cs, err := ipvsconfig.ChangeSet(newconfig)
	if err != nil {
		return &IPVSApplyError{ what: "Unable to build change set from new configuration", origErr: err}
	}

	log.WithField("changeset", cs).Debug("Applying changeset")

	return ipvsconfig.ApplyChangeSet(newconfig, cs, allowedActions)
}

// ApplyChangeSet takes a chhange set and applies all change items to
// the given IPVSConfig 
func (ipvsconfig *IPVSConfig) ApplyChangeSet(newconfig *IPVSConfig, cs *ChangeSet, allowedActions ApplyActions) error {

	ipvs, err := ipvs.New("")
	if err != nil {
		return &IPVSHandleError{}
	}
	defer ipvs.Close()

	// check before hand wether all change set items are covered within allowedActions
	for idx, csiIntf := range cs.Items {
		csi := csiIntf.(ChangeSetItem)
		log.WithFields(log.Fields{
			"idx": idx, 
			"csi": csi,
		}).Tracef("Checking change set item")

		switch csi.Type {
		case DeleteService:
			allowed, found := allowedActions[ApplyActionDeleteService]
			if !found || ! allowed {
				return &IPVSApplyError{what: "not allowed to delete a service"}
			}
		case AddService:
			allowed, found := allowedActions[ApplyActionAddService]
			if !found || ! allowed {
				return &IPVSApplyError{what: "not allowed to add a service"}
			}
			// if service has destinations, check as well if allowed
			if len(csi.Service.Destinations) > 0 {
				allowed, found = allowedActions[ApplyActionAddDestination]
				if !found || ! allowed {
					return &IPVSApplyError{what: "not allowed to add a destinations"}
				}
			}
		case UpdateService:
			allowed, found := allowedActions[ApplyActionUpdateService]
			if !found || ! allowed {
				return &IPVSApplyError{what: "not allowed to update a service"}
			}
		case AddDestination:
			allowed, found := allowedActions[ApplyActionAddDestination]
			if !found || ! allowed {
				return &IPVSApplyError{what: "not allowed to add a destination"}
			}
		case DeleteDestination:
			allowed, found := allowedActions[ApplyActionDeleteDestination]
			if !found || ! allowed {
				return &IPVSApplyError{what: "not allowed to delete a destination"}
			}
		case UpdateDestination:
			allowed, found := allowedActions[ApplyActionUpdateDestination]
			if !found || ! allowed {
				return &IPVSApplyError{what: "not allowed to update a destination"}
			}
		default:
			log.WithField("type", csi.Type).Tracef("Unhandled change type")			
		}
	}

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
