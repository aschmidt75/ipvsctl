package integration

import (
	"testing"
)

func TestSplitProtoHostPort(t *testing.T) {
	tables := []struct {
		in       string
		proto    string
		hostport string
	}{
		{"tcp://1.2.3.4:80/", "tcp", "1.2.3.4:80"},
		{"tcp://1.2.3.4", "tcp", "1.2.3.4"},
		{"udp://some.host/", "udp", "some.host"},
	}

	for _, table := range tables {
		p, hp, err := splitProtoHostPort(table.in)
		if err != nil {
			t.Errorf("Internal error occurred: %s", err)
		}
		if p != table.proto {
			t.Errorf("Proto was incorrect: %s", p)
		}
		if hp != table.hostport {
			t.Errorf("HostPort was incorrect: %s", hp)
		}
	}
}

func TestSplitHostPort(t *testing.T) {
	tables := []struct {
		in   string
		host string
		port int
	}{
		{"1.2.3.4", "1.2.3.4", 0},
		{"1.2.3.4:80", "1.2.3.4", 80},
		{"some.host:443", "some.host", 443},
	}

	for _, table := range tables {
		h, p, err := splitHostPort(table.in)
		if err != nil {
			t.Errorf("Internal error occurred: %s", err)
		}
		if p != table.port {
			t.Errorf("port was incorrect: %d", p)
		}
		if h != table.host {
			t.Errorf("Host was incorrect: %s", h)
		}
	}
}

func TestSplitCompoundAddress(t *testing.T) {
	tables := []struct {
		in           string
		p, ap        string
		port, fwmark int
	}{
		{"fwmark:37", "", "", 0, 37},
		{"tcp://1.2.3.4:80/", "tcp", "1.2.3.4", 80, 0},
		{"tcp://some.host/", "tcp", "some.host", 0, 0},
	}
	for _, table := range tables {
		p, ap, port, fwmark, err := splitCompoundAddress(table.in)
		if err != nil {
			t.Errorf("Internal error occurred: %s", err)
		}
		if p != table.p {
			t.Errorf("Proto was incorrect: %s", p)
		}
		if ap != table.ap {
			t.Errorf("AddressPart was incorrect: %s", ap)
		}
		if port != table.port {
			t.Errorf("port was incorrect: %d", port)
		}
		if fwmark != table.fwmark {
			t.Errorf("fwmark was incorrect: %d", fwmark)
		}
	}

	var invalidIns = []string{"fwmark:abc", "nosuchproto://1.2.3"}
	for _, in := range invalidIns {
		_, _, _, _, err := splitCompoundAddress(in)
		if err == nil {
			t.Errorf("should have produced an error, but did not: %s", in)
		}

	}
}
