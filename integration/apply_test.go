package integration_test

import (
	"io"
	"log"
	"os"
	"testing"

	integration "github.com/aschmidt75/ipvsctl/integration"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
)

func clearIPVS() {
	const targetModel string = "{}"
	var newConfig integration.IPVSConfig
	_ = yaml.Unmarshal([]byte(targetModel), &newConfig)

	opts := integration.ApplyOpts{
		KeepWeights:    true,
		AllowedActions: integration.AllApplyActions(),
	}

	currentConfig := integration.NewIPVSConfig()
	if err := currentConfig.Apply(&newConfig, opts); err != nil {
		panic(err)
	}

}

var (
	l *log.Logger
)

func TestMain(m *testing.M) {
	//l = log.Default()
	l = log.New(io.Discard, "", 0)
	os.Exit(m.Run())
}

func TestApplyGetOnEmptyModel(t *testing.T) {

	const targetModel string = "{}"
	var newConfig integration.IPVSConfig
	_ = yaml.Unmarshal([]byte(targetModel), &newConfig)

	opts := integration.ApplyOpts{
		KeepWeights:    true,
		AllowedActions: integration.AllApplyActions(),
	}

	currentConfig := integration.NewIPVSConfigWithLogger(l)
	if err := currentConfig.Get(); err != nil {
		t.Errorf("Unable to get current ipvs table: %w\n", err)
		t.FailNow()
	}
	if err := currentConfig.Apply(&newConfig, opts); err != nil {
		t.Errorf("Unable to apply test model: %w\n", err)
		t.FailNow()
	}

	updatedConfig := integration.NewIPVSConfigWithLogger(l)
	if err := updatedConfig.Get(); err != nil {
		t.Errorf("Unable to get current ipvs table: %w\n", err)
		t.FailNow()
	}

	/*	fmt.Printf("Current ipvs table:\n")
		for _, s := range updatedConfig.Services {
			fmt.Printf("+ %s (%s)\n", s.Address, s.SchedName)
			for _, d := range s.Destinations {
				fmt.Printf(" -> %s (%s) w:%d\n", d.Address, d.Forward, d.Weight)
			}
		}
	*/
	assert.Len(t, updatedConfig.Services, 0, "Must not have any services")
}

func TestApplyGetOnServices(t *testing.T) {

	clearIPVS()

	const targetModel string = `services:
- address: tcp://127.0.0.1:5678
  sched: rr
- address: tcp://127.0.0.1:1234
  sched: wrr
`
	var newConfig integration.IPVSConfig
	err := yaml.Unmarshal([]byte(targetModel), &newConfig)
	if err != nil {
		panic(err)
	}

	opts := integration.ApplyOpts{
		KeepWeights:    true,
		AllowedActions: integration.AllApplyActions(),
	}

	currentConfig := integration.NewIPVSConfigWithLogger(l)
	if err := currentConfig.Get(); err != nil {
		t.Errorf("Unable to get current ipvs table: %w\n", err)
		t.FailNow()
	}
	if err := currentConfig.Apply(&newConfig, opts); err != nil {
		t.Errorf("Unable to apply test model: %w\n", err)
		t.FailNow()
	}

	updatedConfig := integration.NewIPVSConfigWithLogger(l)
	if err := updatedConfig.Get(); err != nil {
		t.Errorf("Unable to get current ipvs table: %w\n", err)
		t.FailNow()
	}

	assert.Len(t, updatedConfig.Services, 2, "Should have services")
	assert.Equal(t, updatedConfig.Services[0].Address, "tcp://127.0.0.1:5678")
	assert.Equal(t, updatedConfig.Services[0].SchedName, "rr")
	assert.Equal(t, updatedConfig.Services[1].Address, "tcp://127.0.0.1:1234")
	assert.Equal(t, updatedConfig.Services[1].SchedName, "wrr")

	clearIPVS()
}
