package integration

import (
	"fmt"

	ipvs "github.com/aschmidt75/ipvsctl/ipvs"
	log "github.com/sirupsen/logrus"
)

func protoNumToStr(service *ipvs.Service) string {
	switch service.Protocol {
	case 17:
		return "udp"
	case 6:
		return "tcp"
	case 132:
		return "sctp"
	default:
		return "N/A"
	}
}

// IPVSHandleError signals that we cannot obtain an ipvs handle
type IPVSHandleError struct{}

func (e *IPVSHandleError) Error() string {
	return "Unable to create IPVS handle. Is the kernel module installed and active?"
}

// IPVSQueryError signal an error when querying data
type IPVSQueryError struct {
	what string
}

func (e *IPVSQueryError) Error() string {
	return fmt.Sprintf("Unable to query IPVS (%s). Is the kernel module installed and active?", e.what)
}

// Get retrieves the current IPVC config with all services and destinations
func Get() (*IPVSConfig, error) {
	log.Debug("Querying ipvs data...")

	ipvs, err := ipvs.New("")
	if err != nil {
		return nil, &IPVSHandleError{}
	}
	log.Debugf("%#v\n", ipvs)

	res := &IPVSConfig{}

	err = getServicesWithDestinations(ipvs, res)
	return res, err
}

func getForward(d *ipvs.Destination) string {
	if d == nil {
		return ""
	}
	switch d.ConnectionFlags {
	case 0x03:
		return "g"
	case 0x02:
		return "i"
	case 0x00:
		return "m"
	default:
		return "?"
	}
}

func getDestinationsForService(ipvs *ipvs.Handle, service *ipvs.Service, s *Service) error {
	//
	dests, err := ipvs.GetDestinations(service)
	if err != nil {
		return &IPVSQueryError{what: "destinations"}
	}

	if dests != nil && len(dests) > 0 {
		s.Destinations = make([]*Destination, len(dests))

		for idx, dest := range dests {
			log.Debugf("%d -> %#v\n", idx, *dest)

			s.Destinations[idx] = &Destination{
				Address: fmt.Sprintf("%s:%d", dest.Address, dest.Port),
				Weight:  dest.Weight,
				Forward: getForward(dest),
			}
		}
	}

	return nil
}

func getServicesWithDestinations(ipvs *ipvs.Handle, res *IPVSConfig) error {
	services, err := ipvs.GetServices()
	if err != nil {
		return &IPVSQueryError{what: "services"}
	}
	log.Debugf("%#v\n", services)
	if services != nil && len(services) > 0 {
		res.Services = make([]*Service, len(services))

		for idx, service := range services {
			service, err = ipvs.GetService(service)
			if err != nil {
				return &IPVSQueryError{what: "service"}
			}

			log.Debugf("%d -> %#v\n", idx, *service)

			var adrStr = ""
			if service.Protocol != 0 {
				protoStr := protoNumToStr(service)
				ipStr := fmt.Sprintf("%s", service.Address)
				var portStr = ""
				if service.Port != 0 {
					portStr = fmt.Sprintf(":%d", service.Port)
				}
				adrStr = fmt.Sprintf("%s://%s%s", protoStr, ipStr, portStr)
			} else {
				adrStr = fmt.Sprintf("fwmark:%d", service.FWMark)
			}

			res.Services[idx] = &Service{
				Address:   adrStr,
				SchedName: service.SchedName,
			}

			err = getDestinationsForService(ipvs, service, res.Services[idx])
			if err != nil {
				return err
			}
		}
	}

	return nil
}
