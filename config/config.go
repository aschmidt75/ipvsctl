package config

import (
	"github.com/caarlos0/env/v6"
)

// Configuration holds all global config entries
type Configuration struct {
	Trace              bool   `env:"IPVSCTL_LOG_TRACE" envDefault:"false"`
	Debug              bool   `env:"IPVSCTL_LOG_DEBUG" envDefault:"false"`
	Verbose            bool   `env:"IPVSCTL_LOG_VERBOSE" envDefault:"false"`
	ParamsHostNetwork  bool   `env:"IPVSCTL_PARAMS_HOST_NETWORK" envDefault:"false"`
	ParamsHostEnv      bool   `env:"IPVSCTL_PARAMS_HOST_ENV" envDefault:"false"`
	ParamsFilesFromEnv string `env:"IPVSCTL_PARAMS_FILE" envDefault:""`
	ParamsFiles        []string
	ParamsURLsFromEnv  string `env:"IPVSCTL_PARAMS_URLS" envDefault:""`
	ParamsURLs         []string
}

var (
	configuration *Configuration
)

// Config retrieves the current configuration
func Config() *Configuration {
	if configuration == nil {
		configuration = &Configuration{}

		// parse env
		if err := env.Parse(configuration); err != nil {
			panic(err)
		}
	}
	return configuration
}
