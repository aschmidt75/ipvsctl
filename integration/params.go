package integration

import (
	dynp "github.com/aschmidt75/go-dynamic-params"
)

// ResolveParams walks over a complete IPVS configuration and resolves all
// dynamic parameters using the ResolverChain. It returns a copy of the configuration
// or an error
func (ipvsconfig *IPVSConfig) ResolveParams(rc dynp.ResolverChain) (*IPVSConfig, error) {
	res := From(ipvsconfig)

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
			res.Services[idx].Destinations[dIdx].Address = d
		}
	}

	return res, nil
}
