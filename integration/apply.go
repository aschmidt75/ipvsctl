package integration

import (
	"fmt"

	ipvs "github.com/aschmidt75/ipvsctl/ipvs"
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

func isActionAllowed(actions ApplyActions, action ApplyActionType) bool {
	allowed, found := actions[action]
	return found && allowed
}

// Apply compares new config to current config, builds a changeset and
// applies the change set items within.
func (ipvsconfig *IPVSConfig) Apply(newconfig *IPVSConfig, opts ApplyOpts) error {

	// create changeset from new configuration
	cs, err := ipvsconfig.ChangeSet(newconfig, opts)
	if err != nil {
		return &IPVSApplyError{what: "Unable to build change set from new configuration", origErr: err}
	}

	ipvsconfig.log.Printf("Applying changeset, %#v\n", cs)

	return ipvsconfig.ApplyChangeSet(newconfig, cs, opts)
}

// ApplyChangeSet takes a change set and applies all change items to
// the given IPVSConfig
func (ipvsconfig *IPVSConfig) ApplyChangeSet(newconfig *IPVSConfig, cs *ChangeSet, opts ApplyOpts) error {

	ipvs, err := ipvs.New("")
	if err != nil {
		return &IPVSHandleError{}
	}
	defer ipvs.Close()

	allowedActions := opts.AllowedActions

	// check before hand wether all change set items are covered within allowedActions
	for _, csiIntf := range cs.Items {
		csi, ok := csiIntf.(ChangeSetItem)
		if !ok {
			return fmt.Errorf("invalid item in change set: %v", csiIntf)
		}

		switch csi.Type {
		case DeleteService:
			if !isActionAllowed(allowedActions, ApplyActionDeleteService) {
				return &IPVSApplyError{what: "not allowed to delete a service"}
			}
		case AddService:
			if !isActionAllowed(allowedActions, ApplyActionAddService) {
				return &IPVSApplyError{what: "not allowed to add a service"}
			}
			// if service has destinations, check as well if allowed
			if len(csi.Service.Destinations) > 0 {
				if !isActionAllowed(allowedActions, ApplyActionAddDestination) {
					return &IPVSApplyError{what: "not allowed to add a destinations"}
				}
			}
		case UpdateService:
			if !isActionAllowed(allowedActions, ApplyActionUpdateService) {
				return &IPVSApplyError{what: "not allowed to update a service"}
			}
		case AddDestination:
			if !isActionAllowed(allowedActions, ApplyActionAddDestination) {
				return &IPVSApplyError{what: "not allowed to add a destination"}
			}
		case DeleteDestination:
			if !isActionAllowed(allowedActions, ApplyActionDeleteDestination) {
				return &IPVSApplyError{what: "not allowed to delete a destination"}
			}
		case UpdateDestination:
			if !isActionAllowed(allowedActions, ApplyActionUpdateDestination) {
				return &IPVSApplyError{what: "not allowed to update a destination"}
			}
		default:
			ipvsconfig.log.Printf("Unhandled change type: %s", csi.Type)
		}
	}

	for _, csiIntf := range cs.Items {
		csi, ok := csiIntf.(ChangeSetItem)
		if !ok {
			return fmt.Errorf("invalid item in change set: %v", csiIntf)
		}

		ipvsconfig.log.Printf("Applying change set item %#v\n", csi)

		switch csi.Type {
		case DeleteService:
			ipvsconfig.log.Printf("Removing service from current config, addr=%s\n", csi.Service.Address)

			err = ipvs.DelService(csi.Service.service)
			if err != nil {
				return &IPVSApplyError{what: "unable to delete service", origErr: err}
			}
		case AddService:
			ipvsconfig.log.Printf("Adding to current config, addr=%s\n", csi.Service.Address)

			newIPVSService, err := newconfig.NewIpvsServiceStruct(csi.Service)
			if err != nil {
				return &IPVSApplyError{what: "unable to add service", origErr: err}
			}

			ipvsconfig.log.Printf("newIPVSService=%#v\n", newIPVSService)

			err = ipvs.NewService(newIPVSService)
			if err != nil {
				return &IPVSApplyError{what: "unable to add ipvs service", origErr: err}
			}

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
			ipvsconfig.log.Printf("Updating service, addr=%s\n", csi.Service.Address)

			newIPVSService, err := newconfig.NewIpvsServiceStruct(csi.Service)
			if err != nil {
				return &IPVSApplyError{what: "unable to edit service", origErr: err}
			}

			err = ipvs.UpdateService(newIPVSService)
			if err != nil {
				return &IPVSApplyError{what: "unable to edit ipvs service", origErr: err}
			}
			ipvsconfig.log.Printf("edited service: %#v\n", newIPVSService)

		case AddDestination:
			ipvsconfig.log.Printf("Adding destination to current config, dest=%s, svc=%s\n", csi.Destination.Address, csi.Service.Address)

			newIPVSDestination, err := newconfig.NewIpvsDestinationStruct(csi.Destination)
			if err != nil {
				return &IPVSApplyError{what: fmt.Sprintf("unable to prepare new destination for service %s", csi.Service.Address), origErr: err}
			}
			err = ipvs.NewDestination(csi.Service.service, newIPVSDestination)
			if err != nil {
				return &IPVSApplyError{what: fmt.Sprintf("unable to add new destination %#v for service %s", newIPVSDestination.Address, csi.Service.Address), origErr: err}
			}

		case DeleteDestination:
			ipvsconfig.log.Printf("Removing destination from current config, dest=%s, svc=%s\n", csi.Destination.Address, csi.Service.Address)

			err = ipvs.DelDestination(csi.Service.service, csi.Destination.destination)
			if err != nil {
				return &IPVSApplyError{what: fmt.Sprintf("unable to delete destination %s for service %s", csi.Destination.Address, csi.Service.Address), origErr: err}
			}

		case UpdateDestination:
			ipvsconfig.log.Printf("Updating destination, dest=%s, svc=%s\n", csi.Destination.Address, csi.Service.Address)

			updateIPVSDestination, err := newconfig.NewIpvsDestinationStruct(csi.Destination)
			if err != nil {
				return &IPVSApplyError{what: fmt.Sprintf("unable to prepare edited destination for service %s", csi.Service.Address), origErr: err}
			}
			ipvsconfig.log.Printf("Updating destination: %#v\n", updateIPVSDestination)
			err = ipvs.UpdateDestination(csi.Service.service, updateIPVSDestination)
			if err != nil {
				return &IPVSApplyError{what: fmt.Sprintf("unable to update destination %#v for service %s", updateIPVSDestination.Address, csi.Service.Address), origErr: err}
			}

		default:
			ipvsconfig.log.Printf("Unhandled change type %s\n", csi.Type)
		}
	}

	return nil
}
