package integration

import (
	dynp "github.com/aschmidt75/go-dynamic-params"

	log "github.com/sirupsen/logrus"
)

// ResolveParams walks over a complete IPVS configuration and resolves all
// dynamic parameters using the ResolverChain. It returns a copy of the configuration
// or an error
func (ipvsconfig *IPVSConfig) ResolveParams(rc dynp.ResolverChain) (*IPVSConfig, error) {
	res := &IPVSConfig{}

	res.Defaults = ipvsconfig.Defaults

	res.Services = make([]*Service, len(ipvsconfig.Services))
	for idx, service := range ipvsconfig.Services {
		res.Services[idx] = &Service{
			SchedName: service.SchedName,
			service:   service.service,
		}

		s, err := dynp.ResolveFromString(service.Address, rc)
		if err != nil {
			return res, err
		}
		if s != service.Address {
			log.WithFields(log.Fields{
				"from": service.Address,
				"to":   s,
			}).Debug("resolved service address")
		}
		res.Services[idx].Address = s

		res.Services[idx].Destinations = make([]*Destination, len(service.Destinations))
		for dIdx, destination := range service.Destinations {
			res.Services[idx].Destinations[dIdx] = &Destination{
				Weight:      destination.Weight,
				Forward:     destination.Forward,
				destination: destination.destination,
			}

			d, err := dynp.ResolveFromString(destination.Address, rc)
			if err != nil {
				return res, err
			}
			if d != destination.Address {
				log.WithFields(log.Fields{
					"from": destination.Address,
					"to":   d,
				}).Debug("resolved destination address")
			}
			res.Services[idx].Destinations[dIdx].Address = d
		}
	}

	return res, nil
}
