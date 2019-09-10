package integration

import (
	"net"

	ipvs "github.com/aschmidt75/ipvsctl/ipvs"

	//"net"
	"errors"
	"regexp"
	"strconv"
	"strings"
	"syscall"
)

// Service describes an IPVS service entry
type Service struct {
	Address      string         `yaml:"address"`
	SchedName    string         `yaml:"sched"`
	Destinations []*Destination `yaml:"destinations,omitempty"`

	service *ipvs.Service
}

// Destination models a real server behind a service
type Destination struct {
	Address string `yaml:"address"`
	Weight  int    `yaml:"weight"`  // weight for weighted forwarders
	Forward string `yaml:"forward"` // forwards as string (g=gatewaying, i=ipip/tunnel, m=masquerade/nat)
}

// IPVSConfig is a single ipvs setup
type IPVSConfig struct {
	Services []*Service `yaml:"services,omitempty"`
}

// IsEqual returns true if both s and b point to the same address
func (s *Service) IsEqual(b *Service) bool {
	//log.Debugf("Compare: %s <-> %s\n", a.Address, b.Address)
	return s.Address == b.Address
}

func splitProtoHostPort(in string) (proto, hostport string, err error) {
	r := regexp.MustCompile(`^(?P<proto>tcp|udp|sctp)://(?P<host>.*)`)
	x := r.FindStringSubmatch(in)
	if x == nil || len(x) != 3 {
		return "", "", errors.New("error splitting: " + in)
	}
	return strings.TrimRight(x[1], "/ "), strings.TrimRight(x[2], "/ "), nil
}

func splitHostPort(in string) (host string, port int, err error) {
	// 1. todo: check for ipv6 addr e.g. [fe80]

	i := strings.LastIndex(in, ":")
	if i == -1 {
		// no ":", no port there
		return in, 0, nil
	}

	a := strings.Split(in, ":")
	if len(a) != 2 {
		return "", 0, errors.New("parse error")
	}
	p, err := strconv.ParseInt(a[1], 10, 32)
	if err != nil {
		return "", 0, err
	}
	return a[0], int(p), nil
}

func splitCompoundAddress(in string) (procotol, addressPart string, port, fwmark int, err error) {
	if strings.HasPrefix(in, "fwmark:") {
		// treat rest as fwmark integer
		a := strings.Split(in, ":")
		if len(a) != 2 {
			return "", "", 0, 0, errors.New("unable to parse fwmark: " + in)
		}
		f, err := strconv.ParseInt(a[1], 10, 32)
		if err != nil {
			return "", "", 0, 0, err
		}
		return "", "", 0, int(f), nil
	}

	proto, hp, err := splitProtoHostPort(in)
	if err != nil {
		return "", "", 0, 0, err
	}

	host, port, err := splitHostPort(hp)
	if err != nil {
		return "", "", 0, 0, err
	}

	return proto, host, port, 0, nil
}

// NewIpvsServiceStruct creates a new ipvs.service struct from model integration.Service
func (s *Service) NewIpvsServiceStruct() (*ipvs.Service, error) {
	proto, host, port, fwmark, err := splitCompoundAddress(s.Address)
	if err != nil {
		return nil, err
	}

	var protoAsNum uint16
	switch proto {
	case "tcp":
		protoAsNum = 6
	case "udp":
		protoAsNum = 17
	case "sctp":
		protoAsNum = 132
	}

	res := &ipvs.Service{
		Protocol:      protoAsNum,
		Address:       net.ParseIP(host),
		Port:          uint16(port),
		FWMark:        uint32(fwmark),
		AddressFamily: syscall.AF_INET,
		SchedName:     s.SchedName,
		PEName:        "",
		Netmask:       0xffffffff,
	}

	return res, nil
}
