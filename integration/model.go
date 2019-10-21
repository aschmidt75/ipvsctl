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
	SchedName    string         `yaml:"sched,omitempty"`
	Destinations []*Destination `yaml:"destinations,omitempty"`

	service *ipvs.Service // underlay from ipvs package
}

// Destination models a real server behind a service
type Destination struct {
	Address string `yaml:"address"`
	Weight  int    `yaml:"weight,omitempty"`  // weight for weighted forwarders
	Forward string `yaml:"forward,omitempty"` // forwards as string (direct, tunnel, nat)

	destination *ipvs.Destination // underlay from ipvs package
}

// Defaults contains default values for various model elements. If set here they can be
// omitted in Services or Destinations
type Defaults struct {
	Port      *int    `yaml:"port,omitempty"`    // default port
	Weight    *int    `yaml:"weight,omitempty"`  // default weight for weighted forwarders
	SchedName *string `yaml:"sched,omitempty"`   // default scheduler
	Forward   *string `yaml:"forward,omitempty"` // default forwards as string (direct, tunnel, nat)
}

// IPVSConfig is a single ipvs setup
type IPVSConfig struct {
	Defaults Defaults  `yaml:"defaults,omitempty"`
	Services []*Service `yaml:"services,omitempty"`
}

// ChangeSet contains a number of change set items
type ChangeSet struct {
	Items []interface{} `yaml:"items,omitempty"`
}

// ChangeSetItemType as type for const names of types of change set items
type ChangeSetItemType string

const (
	// AddService adds a new service 
	AddService    ChangeSetItemType = "add-service"

	// UpdateService edits an existing service
	UpdateService ChangeSetItemType = "update-service"

	// DeleteService deletes an existing service
	DeleteService ChangeSetItemType = "delete-service"

	// AddDestination adds a new destination to an existing service
	AddDestination    ChangeSetItemType = "add-destination"

	// UpdateDestination edits an existing destination
	UpdateDestination ChangeSetItemType = "update-destination"

	// DeleteDestination deletes an existing destination
	DeleteDestination ChangeSetItemType = "delete-destination"
)

// ChangeSetItem ...
type ChangeSetItem struct {
	Type        ChangeSetItemType `yaml:"type"`
	Description string
	Service     *Service          `yaml:"service,omitempty"`
	Destination *Destination      `yaml:"destination,omitempty"`
}

type ApplyOpts struct {
	KeepWeights bool
}

// NewChangeSet makes a new changeset
func NewChangeSet() *ChangeSet {
	return &ChangeSet{
		Items: make([]interface{}, 0, 5),
	}
}

// AddChange adds a new item to the changeset
func (cs *ChangeSet) AddChange(csi ChangeSetItem) {
	cs.Items = append(cs.Items, csi)
}

// IsEqual for Services returns true if both s and b point to the same address string (that includes the protocol)
func (s *Service) IsEqual(b *Service) bool {
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
		return "", 0, errors.New("parse error in "+in)
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
func (c *IPVSConfig) NewIpvsServiceStruct(s *Service) (*ipvs.Service, error) {
	proto, host, port, fwmark, err := splitCompoundAddress(s.Address)
	if err != nil {
		return nil, err
	}
	if port == 0 {
		if c.Defaults.Port != nil {
			port = *c.Defaults.Port
		}
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

	schedName := s.SchedName
	if schedName == "" {
		if c.Defaults.SchedName != nil && *c.Defaults.SchedName != "" {
			schedName = *c.Defaults.SchedName
		} else {
			schedName = "rr"
		}
	}

	res := &ipvs.Service{
		Protocol:      protoAsNum,
		Address:       net.ParseIP(host),
		Port:          uint16(port),
		FWMark:        uint32(fwmark),
		AddressFamily: syscall.AF_INET,
		SchedName:     schedName,
		PEName:        "",
		Netmask:       0xffffffff,
	}

	return res, nil
}

// NewIpvsDestinationsStruct creates a new ipvs.Destination struct array from model integration.Service
func (c *IPVSConfig) NewIpvsDestinationsStruct(s *Service) ([]*ipvs.Destination, error) {
	res := make([]*ipvs.Destination, len(s.Destinations))

	for idx, destination := range s.Destinations {
		var err error
		res[idx], err = c.NewIpvsDestinationStruct(destination)
		if err != nil {
			return nil, err
		}
	}

	return res, nil
}

// NewIpvsDestinationStruct creates a single new ipvs.Destination struct from model integration.Service
func (c *IPVSConfig) NewIpvsDestinationStruct(destination *Destination) (*ipvs.Destination, error) {
	h, p, err := splitHostPort(destination.Address)
	if err != nil {
		return nil, err
	}
	if p == 0 {
		if c.Defaults.Port != nil {
			p = *c.Defaults.Port
		}
	}

	df := destination.Forward
	if df == "" {
		if c.Defaults.Forward != nil {
			df = *c.Defaults.Forward
		}
	}
	var cf uint32
	switch df {
	case "direct":
		cf = 0x3
	case "tunnel":
		cf = 0x2
	case "nat":
		cf = 0x0
	default:
		return nil, errors.New("bad forward. Must be one of direct, tunnel or nat")
	}

	w := destination.Weight
	if w == 0 {
		if c.Defaults.Weight != nil {
			w = *c.Defaults.Weight
		}
	}
	if w < 0 || w > 65535 {
		w = 1
	}
	return &ipvs.Destination{
		Address:         net.ParseIP(h),
		Port:            uint16(p),
		ConnectionFlags: cf,
		Weight:          w,
	}, nil
}

// CompareServicesEquality for Service does a complete compare. it applies config defaults
func CompareServicesEquality(ca *IPVSConfig, a *Service, cb *IPVSConfig, b *Service) (bool, error) {
	var err error

	ident, err := CompareServicesIdentifyingEquality(ca, a, cb, b )
	if err != nil {
		return false, err
	}
	if ident == false {
		return false, nil
	}

	// compare Scheduler
	af := a.SchedName
	bf := b.SchedName
	if af == "" && ca.Defaults.SchedName != nil && *ca.Defaults.SchedName != "" {
		af = *ca.Defaults.SchedName
	}
	if bf == "" && cb.Defaults.SchedName != nil && *cb.Defaults.SchedName != "" {
		bf = *cb.Defaults.SchedName
	}
	if af == "" {
		af = "rr"
	}
	if bf == "" {
		bf = "rr"
	}
	if af != bf {
		return false, nil
	}

	// everything is equal
	return true, nil
}

// CompareServicesIdentifyingEquality for Service compares service address including defaults
func CompareServicesIdentifyingEquality(ca *IPVSConfig, a *Service, cb *IPVSConfig, b *Service) (bool, error) {
	var err error

	apr, ah, ap, afwm, err := splitCompoundAddress(a.Address)
	if err != nil {
		return false, err
	}
	if ap == 0 && ca.Defaults.Port != nil && *ca.Defaults.Port != 0 {
		ap = *ca.Defaults.Port
	}
	bpr, bh, bp, bfwm, err := splitCompoundAddress(b.Address)
	if err != nil {
		return false, err
	}
	if bp == 0 && cb.Defaults.Port != nil && *cb.Defaults.Port != 0 {
		bp = *cb.Defaults.Port
	}

	if afwm != 0 && bfwm != 0 && afwm != bfwm {
		// both use fwmark, but different one
		return false, nil
	}
	if apr != bpr {
		return false, nil
	}
	if ah != bh {
		return false, nil
	}
	if ap != bp {
		return false, nil
	}

	// everything is equal
	return true, nil
}

func CompareDestinationsEquality(ca *IPVSConfig, a *Destination, cb *IPVSConfig, b *Destination, opts AppyOpts) (bool, error) {
	var err error

	// compare host+port
	ah, ap, err := splitHostPort(a.Address)
	if err != nil {
		return false, err
	}
	if ap == 0 && ca.Defaults.Port != nil && *ca.Defaults.Port != 0 {
		ap = *ca.Defaults.Port
	}

	bh, bp, err := splitHostPort(b.Address)
	if err != nil {
		return false, err
	}
	if bp == 0 && cb.Defaults.Port != nil && *cb.Defaults.Port != 0 {
		bp = *cb.Defaults.Port
	}

	if ah != bh {
		return false, nil
	}
	if ap != bp {
		return false, nil
	}

	// compare forward
	af := a.Forward
	bf := b.Forward
	if af == "" && ca.Defaults.Forward != nil && *ca.Defaults.Forward != "" {
		af = *ca.Defaults.Forward
	}
	if bf == "" && cb.Defaults.Forward != nil && *cb.Defaults.Forward != "" {
		bf = *cb.Defaults.Forward
	}
	if af != bf {
		return false, nil
	}

	if opts.KeepWeights == false {
		// compare weight
		aw := a.Weight
		bw := b.Weight
		if aw == 0 && ca.Defaults.Weight != nil && *ca.Defaults.Weight != 0 {
			aw = *ca.Defaults.Weight
		}
		if bw == 0 && cb.Defaults.Weight != nil && *cb.Defaults.Weight != 0 {
			bw = *cb.Defaults.Weight
		}
		// default weight is 1
		if aw == 0 {
			aw = 1
		}
		if bw == 0 {
			bw = 1
		}

		if aw != bw {
			return false, nil
		}
	}

	// everything is equal
	return true, nil
}

func CompareDestinationIdentifyingEquality(ca *IPVSConfig, a *Destination, cb *IPVSConfig, b *Destination) (bool, error) {
	var err error

	// compare host+port
	ah, ap, err := splitHostPort(a.Address)
	if err != nil {
		return false, err
	}
	if ap == 0 && ca.Defaults.Port != nil && *ca.Defaults.Port != 0 {
		ap = *ca.Defaults.Port
	}

	bh, bp, err := splitHostPort(b.Address)
	if err != nil {
		return false, err
	}
	if bp == 0 && cb.Defaults.Port != nil && *cb.Defaults.Port != 0 {
		bp = *cb.Defaults.Port
	}

	if ah != bh {
		return false, nil
	}
	if ap != bp {
		return false, nil
	}
	// everything is equal
	return true, nil
}

func (ipvs *IPVSConfig) LocateServiceAndDestination(serviceHandle, destinationHandle string) (*Service, *Destination) {
	var s *Service = nil
	var d *Destination = nil

	for _, service := range ipvs.Services {
		if service.service == nil {
			continue
		}
		a := MakeAdressStringFromIpvsService(service.service)

		if a == serviceHandle {
			s = service

			// 
			for _, destination := range service.Destinations {
				if destination.destination == nil {
					continue
				}
				b := MakeAdressStringFromIpvsDestination(destination.destination)
				if b == destinationHandle {
					d = destination
					break
				}
			}

			break
		}
	}

	return s, d
}