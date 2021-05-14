package cmd

import (
	"testing"

	integration "github.com/aschmidt75/ipvsctl/integration"
	"github.com/stretchr/testify/assert"
)

func TestAllowedActions(t *testing.T) {
	var tests = []struct {
		inp string
		res integration.ApplyActions
	}{
		{"*", integration.ApplyActions{
			integration.ApplyActionAddService:        true,
			integration.ApplyActionUpdateService:     true,
			integration.ApplyActionDeleteService:     true,
			integration.ApplyActionAddDestination:    true,
			integration.ApplyActionUpdateDestination: true,
			integration.ApplyActionDeleteDestination: true,
		}},
	}

	for _, test := range tests {
		t.Run(test.inp, func(t *testing.T) {
			res, err := parseAllowedActions(&test.inp)
			if err != nil {
				t.Error("Should have passed but returned error: %w", err)
			}
			assert.EqualValuesf(t, res, test.res, "Compare")
		})
	}

	invalidActions := "as,nosuchaction,ds"
	_, err := parseAllowedActions(&invalidActions)
	assert.Error(t, err)

	_, err = parseAllowedActions(nil)
	assert.Error(t, err)
}
