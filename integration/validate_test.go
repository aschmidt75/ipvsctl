package integration_test

import (
	"testing"

	integration "github.com/aschmidt75/ipvsctl/integration"
	"gopkg.in/yaml.v2"
)

func TestValidate(t *testing.T) {

	var tests = []struct {
		model string
		ok    bool
	}{
		{`{}`, true},
		{`
services:
- address: tcp://127.0.0.1:9876
  sched: rr
  destinations:
  - address: 127.0.0.2:1234
    weight: 100
    forward: nat
  - address: 127.0.0.3:1234
    weight: 100
    forward: nat
- address: tcp://127.0.0.5:9876
  sched: rr
`, true},
		{`services:
- address: tcp://127.0.0.1:9876
  sched: nosuchsched
`, false},
		{`services:
- address: tcp://not.an.ip:9876
  sched: rr
`, false},
		{`services:
- sched: rr
`, false},
		{`
services:
- address: tcp://127.0.0.1:9876
  sched: rr
- address: tcp://127.0.0.1:9876
  sched: rr
`, false},
		{`
services:
- address: not://anaddre:ss
  sched: rr
`, false},
		{`
services:
- address: fwmark://abcdef
  sched: rr
`, false},
		{`
services:
- address: tcp://127.0.0.1:9876
  sched: rr
  destinations.
  - weight: 100
`, false},
		{`
services:
- address: tcp://127.0.0.1:9876
  sched: rr
  destinations.
  - address: 127.0.0.2:1234
    weight: -100
`, false},
		{`
services:
- address: tcp://127.0.0.1:9876
  sched: rr
  destinations.
  - address: not.an.address:abcd
    weight: 100
`, false},
		{`
services:
- address: tcp://127.0.0.1:9876
  sched: rr
  destinations.
  - address: 127.0.0.2:99887766
    weight: 100
`, false},
		{`
services:
- address: tcp://127.0.0.1:9876
  sched: rr
  destinations.
  - address: 127.0.0.2:123
    weight: 100
	forward: nosuchforward
`, false},
	}

	for _, test := range tests {
		t.Run(test.model, func(t *testing.T) {
			err := validate(t, test.model)
			if err == nil {
				if test.ok == false {
					t.Error("Should have returned a validation error, but did not")
				}
			} else {
				if test.ok == true {
					t.Error("Should have passed but returned a validation error: %w", err)

				}
			}
		})
	}
}

func TestValidateWithDefaults(t *testing.T) {

	var tests = []struct {
		model string
		ok    bool
	}{
		{`
defaults:
  port: 1234
  weight: 100
  sched: rr
  forward: nat
services:
- address: tcp://127.0.0.1:9876
  destinations:
  - address: 127.0.0.2:1234
`, true},
		{`
defaults:
  port: 99887766
  weight: 100
  sched: rr
  forward: nat
`, false},
		{`
defaults:
  port: 1234
  weight: -100
  sched: rr
  forward: nat
`, false},
		{`
defaults:
  port: 1234
  weight: 100
  sched: nosuchsched
  forward: nat
`, false},
		{`
defaults:
  port: 1234
  weight: 100
  sched: rr
  forward: nosuchforward
`, false},
	}

	for _, test := range tests {
		t.Run(test.model, func(t *testing.T) {
			err := validate(t, test.model)
			if err == nil {
				if test.ok == false {
					t.Error("Should have returned a validation error, but did not")
				}
			} else {
				if test.ok == true {
					t.Error("Should have passed but returned a validation error: %w", err)

				}
			}
		})
	}
}

func validate(t *testing.T, model string) error {
	var err error
	var config integration.IPVSConfig
	if err = yaml.Unmarshal([]byte(model), &config); err != nil {
		return err
	}

	return config.Validate()
}
