package integration_test

import (
	"testing"

	dynp "github.com/aschmidt75/go-dynamic-params"
	integration "github.com/aschmidt75/ipvsctl/integration"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
)

func TestParams(t *testing.T) {

	m := map[string]string{
		"h": "127.0.0.1",
		"p": "8080",
	}
	r := dynp.NewMapResolver().
		With(m)

	rc := dynp.ResolverChain{r}

	var tests = []struct {
		model string
		ok    bool
	}{
		{`{}`, true},
		{`
services:
- address: ${h}:${p}
`, true},
		{`
services:
- address: ${h}:${nosuchparam}
`, false},
		{`
services:
- address: tpc://127.0.0.1:8080
  destinations:
  - address: ${h}:${p}
`, true},
		{`
services:
- address: tpc://127.0.0.1:8080
  destinations:
  - address: ${h}:${nosuchparam}
`, false},
	}

	for _, test := range tests {
		t.Run(test.model, func(t *testing.T) {
			var err error
			var baseConfig integration.IPVSConfig
			if err = yaml.Unmarshal([]byte(test.model), &baseConfig); err != nil {
				assert.Nil(t, err)
			}

			resolvedConfig, err := baseConfig.ResolveParams(rc)
			if test.ok {
				assert.Nil(t, err)
				assert.NotNil(t, resolvedConfig)
			} else {
				assert.NotNil(t, err)
			}
		})
	}
}
