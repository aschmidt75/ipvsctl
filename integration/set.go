package integration

import (
	"fmt"
	"time"
)

// IPVSetError signals an error when updating configuration
type IPVSetError struct {
	what    string
	origErr error
}

func (e *IPVSetError) Error() string {
	if e.origErr == nil {
		return fmt.Sprintf("Unable to set new value: %s", e.what)
	}
	return fmt.Sprintf("Unable to set new value: %s\nReason: %s", e.what, e.origErr)
}

// SetWeight sets the destination's weight to newWeight
func (ipvsconfig *IPVSConfig) SetWeight(serviceName, destinationName string, newWeight int) error {
	s, d := ipvsconfig.LocateServiceAndDestination(serviceName, destinationName)
	if s == nil {
		return &IPVSetError{what: fmt.Sprintf("Service %s not found in active ipvs configuration. Try ipvsctl get\n", serviceName)}
	}
	if d == nil {
		return &IPVSetError{what: fmt.Sprintf("Destination %s not found in active ipvs configuration. Try ipvsctl get\n", destinationName)}
	}

	// tweak
	d.Weight = newWeight

	// create changeset
	cs := NewChangeSet()
	cs.AddChange(ChangeSetItem{
		Type:        UpdateDestination,
		Service:     s,
		Destination: d,
	})

	ipvsconfig.log.Printf("applying changeset %s\n", cs)

	err := ipvsconfig.ApplyChangeSet(ipvsconfig, cs, ApplyOpts{
		AllowedActions: ApplyActions{
			ApplyActionUpdateDestination: true,
		}})
	if err == nil {
		ipvsconfig.log.Printf("Updated weight to %d for %s/%s\n", d.Weight, s.Address, d.Address)
	}
	return err
}

const (
	// ControlAdvance advances the time ticker (see cmd/set.go)
	ControlAdvance = 1
	// ControlExit immediately exists the loop
	ControlExit = 2
	// ControlFinish finishes the loop
	ControlFinish = 3
)

// ContinousControlCh is a signalling channel
type ContinousControlCh chan int

// SetWeightContinuous sets the weight of a destination to a target value, within a
// given amount of time, controlled by a channel.
func (ipvsconfig *IPVSConfig) SetWeightContinuous(
	serviceName, destinationName string,
	toWeight int,
	amountOfTimeSecs int,
	cch ContinousControlCh) error {

	if amountOfTimeSecs <= 1 {
		return ipvsconfig.SetWeight(serviceName, destinationName, toWeight)
	}

	s, d := ipvsconfig.LocateServiceAndDestination(serviceName, destinationName)
	if s == nil {
		return &IPVSetError{what: fmt.Sprintf("Service %s not found in active ipvs configuration. Try ipvsctl get\n", serviceName)}
	}
	if d == nil {
		return &IPVSetError{what: fmt.Sprintf("Destination %s not found in active ipvs configuration. Try ipvsctl get\n", destinationName)}
	}

	fromWeight := d.Weight
	cs := NewChangeSet()
	cs.AddChange(ChangeSetItem{
		Type:        UpdateDestination,
		Service:     s,
		Destination: d,
	})

	// get time now
	timeStart := time.Now()

	for {
		// wait for command
		cmd := <-cch

		switch cmd {
		case ControlExit:
			return nil

		case ControlFinish:
			return ipvsconfig.SetWeight(serviceName, destinationName, toWeight)

		case ControlAdvance:
			timeElapsed := time.Now().Sub(timeStart)
			if timeElapsed > 0 {
				percElapsed := timeElapsed.Seconds() / float64(amountOfTimeSecs)
				if percElapsed >= 1 {
					percElapsed = 1
				}
				d.Weight = int(float64(fromWeight) + float64(toWeight-fromWeight)*percElapsed)

				err := ipvsconfig.ApplyChangeSet(ipvsconfig, cs, ApplyOpts{
					AllowedActions: ApplyActions{
						ApplyActionUpdateDestination: true,
					}})
				if err != nil {
					return err
				}
				//ipvsconfig.log.Printf("applying changeset %#v\n", cs)
				ipvsconfig.log.Printf("Updated weight %d [elapsed %d %]\n", d.Weight, int(100*percElapsed))

			}
		}
	}
}
