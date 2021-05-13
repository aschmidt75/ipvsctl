package integration_test

import (
	"testing"

	integration "github.com/aschmidt75/ipvsctl/integration"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
)

func TestChangeSetDestinations(t *testing.T) {

	genmsg := "Unable to build changeset, but should have been: %w\n"

	var cs *integration.ChangeSet
	var err error

	// add destination
	cs, err = buildChangeSet(t, `
services:
- address: tcp://127.0.0.1:9876
  sched: rr
`, `
services:
- address: tcp://127.0.0.1:9876
  sched: rr
  destinations:
  - address: 127.0.0.2:1234
    weight: 100
    forward: nat
`)

	if err != nil {
		t.Errorf(genmsg, err)
	}
	assert.Len(t, cs.Items, 1, "Check changeset item count")

	item := cs.Items[0].(integration.ChangeSetItem)
	assert.Equal(t, item.Type, integration.AddDestination)
	assert.NotNil(t, item.Service)
	assert.NotNil(t, item.Destination)
	assert.Equal(t, item.Destination.Address, "127.0.0.2:1234")
	assert.Equal(t, item.Destination.Weight, 100)
	assert.Equal(t, item.Destination.Forward, "nat")

	// delete destination
	cs, err = buildChangeSet(t, `
services:
- address: tcp://127.0.0.1:9876
  sched: rr
  destinations:
  - address: 127.0.0.2:1234
    weight: 100
    forward: nat
`, `
services:
- address: tcp://127.0.0.1:9876
  sched: rr
`)

	if err != nil {
		t.Errorf(genmsg, err)
	}
	assert.Len(t, cs.Items, 1, "Check changeset item count")

	item = cs.Items[0].(integration.ChangeSetItem)
	assert.Equal(t, item.Type, integration.DeleteDestination)

	// update destination
	cs, err = buildChangeSet(t, `
services:
- address: tcp://127.0.0.1:9876
  sched: rr
  destinations:
  - address: 127.0.0.2:1234
    weight: 100
    forward: nat
`, `
services:
- address: tcp://127.0.0.1:9876
  sched: rr
  destinations:
  - address: 127.0.0.2:1234
    weight: 200
    forward: nat
`)

	if err != nil {
		t.Errorf(genmsg, err)
	}
	assert.Len(t, cs.Items, 1, "Check changeset item count")

	item = cs.Items[0].(integration.ChangeSetItem)
	assert.Equal(t, item.Type, integration.UpdateDestination)
	assert.Equal(t, item.Destination.Weight, 200) // TODO: add additional test where keepWeights = true, so it does not change

	// mixed test
	cs, err = buildChangeSet(t, `
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
`, `
services:
- address: tcp://127.0.0.1:9876
  sched: rr
  destinations:
  - address: 127.0.0.3:1234
    weight: 200
    forward: nat
  - address: 127.0.0.5:1234
    weight: 200
    forward: nat
`)

	if err != nil {
		t.Errorf(genmsg, err)
	}
	assert.Len(t, cs.Items, 3, "Check changeset item count")

	typeMap := make(map[integration.ChangeSetItemType]int)

	for _, item := range cs.Items {
		csitem := item.(integration.ChangeSetItem)
		typeMap[csitem.Type]++
	}

	assert.Equal(t, typeMap[integration.AddDestination], 1)
	assert.Equal(t, typeMap[integration.UpdateDestination], 1)
	assert.Equal(t, typeMap[integration.DeleteDestination], 1)
}

func TestChangeSetServices(t *testing.T) {

	genmsg := "Unable to build changeset, but should have been: %w\n"

	cs, err := buildChangeSet(t, "{}", "{}")
	if err != nil {
		t.Errorf(genmsg, err)
	}
	assert.Len(t, cs.Items, 0, "ChangeSet must be empty")

	// add service w/ destination
	cs, err = buildChangeSet(t, "{}", `
services:
- address: tcp://127.0.0.1:9876
  sched: rr
  destinations:
  - address: 127.0.0.2:1234
    weight: 100
    forward: nat
`)

	if err != nil {
		t.Errorf(genmsg, err)
	}
	assert.Len(t, cs.Items, 1, "Check changeset item count")

	item := cs.Items[0].(integration.ChangeSetItem)
	assert.Equal(t, item.Type, integration.AddService)
	assert.NotNil(t, item.Service)
	assert.Nil(t, item.Destination)
	assert.Equal(t, item.Service.Address, "tcp://127.0.0.1:9876")
	assert.Equal(t, item.Service.SchedName, "rr")
	assert.NotNil(t, item.Service.Destinations)
	assert.Len(t, item.Service.Destinations, 1)
	assert.Equal(t, item.Service.Destinations[0].Address, "127.0.0.2:1234")
	assert.Equal(t, item.Service.Destinations[0].Weight, 100)
	assert.Equal(t, item.Service.Destinations[0].Forward, "nat")

	// update service
	cs, err = buildChangeSet(t, `
services:
- address: tcp://127.0.0.1:9876
  sched: rr
`, `
services:
- address: tcp://127.0.0.1:9876
  sched: wrr
`)

	if err != nil {
		t.Errorf(genmsg, err)
	}
	assert.Len(t, cs.Items, 1, "Check changeset item count")

	item = cs.Items[0].(integration.ChangeSetItem)
	assert.Equal(t, item.Type, integration.UpdateService)
	assert.Equal(t, item.Service.SchedName, "wrr")

	// delete service
	cs, err = buildChangeSet(t, `
services:
- address: tcp://127.0.0.1:9876
  sched: rr
`, ``)

	if err != nil {
		t.Errorf(genmsg, err)
	}
	assert.Len(t, cs.Items, 1, "Check changeset item count")

	item = cs.Items[0].(integration.ChangeSetItem)
	assert.Equal(t, item.Type, integration.DeleteService)

	// multi example
	cs, err = buildChangeSet(t, `
services:
- address: tcp://127.0.0.1:9876
  sched: rr
- address: tcp://127.0.0.2:9876
  sched: rr
`, `
services:
- address: tcp://127.0.0.2:9876
  sched: wrr
- address: tcp://127.0.0.3:9876
  sched: rr
`)

	if err != nil {
		t.Errorf(genmsg, err)
	}

	assert.Len(t, cs.Items, 3, "Check changeset item count")
	typeMap := make(map[integration.ChangeSetItemType]int)

	for _, item := range cs.Items {
		csitem := item.(integration.ChangeSetItem)
		typeMap[csitem.Type]++
	}

	assert.Equal(t, typeMap[integration.AddService], 1)
	assert.Equal(t, typeMap[integration.UpdateService], 1)
	assert.Equal(t, typeMap[integration.DeleteService], 1)
}

func buildChangeSet(t *testing.T, baseModel, changeModel string) (*integration.ChangeSet, error) {
	var err error
	var baseConfig, changeConfig integration.IPVSConfig
	if err = yaml.Unmarshal([]byte(baseModel), &baseConfig); err != nil {
		return nil, err
	}
	if err = yaml.Unmarshal([]byte(changeModel), &changeConfig); err != nil {
		return nil, err
	}

	opts := integration.ApplyOpts{
		KeepWeights:    false,
		AllowedActions: integration.AllApplyActions(),
	}

	cs, err := baseConfig.ChangeSet(&changeConfig, opts)
	if err != nil {
		return nil, err
	}

	t.Logf("cs: %+v\n", cs)

	return cs, nil
}
