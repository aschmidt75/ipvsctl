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
	schedNames = []string{"rr", "wrr", "lc", "wlc", "lblc", "lblcr", "dh", "sh", "sed", "nq"}
)

// Validate checks ipvsconfig for structural errors
func (ipvsconfig *IPVSConfig) Validate() error {

	for _, service := range ipvsconfig.Services {
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

	return nil
}
