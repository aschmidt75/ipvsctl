package model

// Service describes an IPVS service entry
type Service struct {
	Address      string         `json:"address" yaml:"address"`
	FWMark       uint32         `json:"fwmark" yaml:"fwmark"`
	SchedName    string         `json:"sched" yaml:"sched"`
	Destinations []*Destination `json:"destinations,omitempty" yaml:"destinations,omitempty"`
}

// Destination models a real server behind a service
type Destination struct {
	Address string `json:"address" yaml:"address"`
	Weight  int    `json:"weight" yaml:"weight"`
}

// IPVSConfig is a single ipvs setup
type IPVSConfig struct {
	Services []*Service `json:"services,omitempty" yaml:"services,omitempty"`
}
