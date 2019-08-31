package integration

// Service describes an IPVS service entry
type Service struct {
	Address      string         `yaml:"address"`
	SchedName    string         `yaml:"sched"`
	Destinations []*Destination `yaml:"destinations,omitempty"`
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
