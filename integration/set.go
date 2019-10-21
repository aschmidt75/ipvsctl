package integration

import (
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
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
	log.WithFields(log.Fields{
		"service":     serviceName,
		"destination": destinationName,
	}).Debugf("Locating")
	s, d := ipvsconfig.LocateServiceAndDestination(serviceName, destinationName)
	if s != nil {
		log.WithField("service", s).Trace("Found service")
	} else {
		return &IPVSetError{what: fmt.Sprintf("Service %s not found in active ipvs configuration. Try ipvsctl get\n", serviceName)}
	}
	if d != nil {
		log.WithField("destination", d).Trace("Found destination")
	} else {
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

	log.WithField("changeset", cs).Tracef("applying changeset.")

	err := ipvsconfig.ApplyChangeSet(ipvsconfig, cs, ApplyOpts{
		AllowedActions: ApplyActions{
			ApplyActionUpdateDestination: true,
		}})
	if err == nil {
		log.Infof("Updated weight to %d for %s/%s", d.Weight, s.Address, d.Address)
	}
	return err
}

const (
	ControlAdvance = 1
	ControlExit    = 2
	ControlFinish  = 3
)

type ContinousControlCh chan int

func (ipvsconfig *IPVSConfig) SetWeightContinuous(
	serviceName, destinationName string,
	toWeight int,
	amountOfTimeSecs int,
	cch ContinousControlCh) error {

	if amountOfTimeSecs <= 1 {
		return ipvsconfig.SetWeight(serviceName, destinationName, toWeight)
	}

	log.WithFields(log.Fields{
		"service":     serviceName,
		"destination": destinationName,
	}).Debugf("Locating")
	s, d := ipvsconfig.LocateServiceAndDestination(serviceName, destinationName)
	if s != nil {
		log.WithField("service", s).Trace("Found service")
	} else {
		return &IPVSetError{what: fmt.Sprintf("Service %s not found in active ipvs configuration. Try ipvsctl get\n", serviceName)}
	}
	if d != nil {
		log.WithField("destination", d).Trace("Found destination")
	} else {
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

				log.WithFields(log.Fields{
					"t": timeElapsed,
					"p": percElapsed,
					"w": d.Weight,
				}).Trace("calculate distance")

				err := ipvsconfig.ApplyChangeSet(ipvsconfig, cs, ApplyOpts{
					AllowedActions: ApplyActions{
						ApplyActionUpdateDestination: true,
					}})
				if err != nil {
					return err
				}
				log.WithField("changeset", cs).Trace("applying changeset.")
				log.WithFields(log.Fields{
					"weight":   d.Weight,
					"elapsed%": int(100 * percElapsed),
				}).Info("Updated weight")

			}
		}
	}
}
