package integration

import (
	"fmt"
	"net"
)

// IPVSValidateError signal an error when validating a configuration
type IPVSValidateError struct {
	What string
}

func (e *IPVSValidateError) Error() string {
	return fmt.Sprintf("Configuration not valid: %s", e.What)
}

var (
	schedNames   = []string{"rr", "wrr", "lc", "wlc", "lblc", "lblcr", "dh", "sh", "sed", "nq"}
	forwardNames = []string{"direct", "nat", "tunnel"}
)

// Validate checks ipvsconfig for structural errors
func (ipvsconfig *IPVSConfig) Validate() error {

	for _, service := range ipvsconfig.Services {
		if service.Address == "" {
			return &IPVSValidateError{What: fmt.Sprintf("Service address may not be empty")}
		}

		//proto, adrpart, port, fwmark, err
		_, adrpart, _, fwmark, err := splitCompoundAddress(service.Address)
		if err != nil {
			es := fmt.Sprintf("unable to parse address (%s). Must be of format <proto>://<host>[:port] or fwmark:<id>.", service.Address)
			return &IPVSValidateError{What: es}
		}

		if fwmark == 0 {
			// check for ip address
			ip := net.ParseIP(adrpart)
			if ip == nil {
				return &IPVSValidateError{What: fmt.Sprintf("unable to parse address (%s). Not an IP address.", adrpart)}
			}

			if !(len(ip) == net.IPv6len &&
				ip[0] == 0x0 &&
				ip[1] == 0x0 &&
				ip[2] == 0x0 &&
				ip[3] == 0x0 &&
				ip[4] == 0x0 &&
				ip[5] == 0x0 &&
				ip[6] == 0x0 &&
				ip[7] == 0x0 &&
				ip[8] == 0x0 &&
				ip[9] == 0x0 &&
				ip[10] == 0xff &&
				ip[11] == 0xff) {
				return &IPVSValidateError{What: fmt.Sprintf("unable to parse address (%s). IPv6 not supported.", adrpart)}
			}
		} else {
			if fwmark < 0 || fwmark > 65535 {
				return &IPVSValidateError{What: fmt.Sprintf("unable to parse address (%s). Invalid fwmark number.", adrpart)}
			}
		}

		// check scheduler if given
		if service.SchedName != "" {
			bOk := false
			for _, sn := range schedNames {
				if sn == service.SchedName {
					bOk = true
					break
				}
			}
			if !bOk {
				return &IPVSValidateError{What: fmt.Sprintf("invalid scheduler (%s) for service (%s).", service.SchedName, service.Address)}
			}
		}

		// check destination addresses
		for _, destination := range service.Destinations {
			if destination.Address == "" {
				return &IPVSValidateError{What: fmt.Sprintf("Destination address may not be empty for service %s", service.Address)}
			}

			h, p, err := splitHostPort(destination.Address)
			if err != nil {
				return &IPVSValidateError{What: fmt.Sprintf("unable to parse address (%s) for service %s. Check host and port.", destination.Address, service.Address)}
			}
			// check for ip address
			ip := net.ParseIP(h)
			if ip == nil {
				return &IPVSValidateError{What: fmt.Sprintf("unable to parse address (%s) for service %s. Not an IP address.", h, service.Address)}
			}
			if p < 1 || p > 65535 {
				return &IPVSValidateError{What: fmt.Sprintf("bad port (%d) for service %s", p, service.Address)}
			}

			if destination.Forward != "" {
				bOk := false
				for _, sn := range forwardNames {
					if sn == destination.Forward {
						bOk = true
					}
				}
				if !bOk {
					return &IPVSValidateError{What: fmt.Sprintf("invalid forward (%s) for destination %s in service %s.", destination.Forward, destination.Address, service.Address)}
				}
			}

			if destination.Weight < 0 || destination.Weight > 65535 {
				return &IPVSValidateError{What: fmt.Sprintf("invalid weight (%d) for destination %s in service %s.", destination.Weight, destination.Address, service.Address)}
			}

		}
	}

	return nil
}
