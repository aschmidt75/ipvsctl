package cmd

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"strings"

	"encoding/json"

	"gopkg.in/yaml.v2"

	dynp "github.com/aschmidt75/go-dynamic-params"
	"github.com/aschmidt75/ipvsctl/config"
	integration "github.com/aschmidt75/ipvsctl/integration"
)

func readInput(filename *string) ([]byte, error) {
	var b []byte
	var err error
	if *filename == "-" {
		b, err = ioutil.ReadAll(os.Stdin)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading from STDIN\n")
			os.Exit(exitInvalidFile)
		}
	} else {
		b, err = ioutil.ReadFile(*filename)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading from input file %s\n", *filename)
			os.Exit(exitInvalidFile)
		}
	}

	return b, err
}

func readModelFromInput(filename *string) (*integration.IPVSConfig, error) {
	c := integration.NewIPVSConfig()

	b, err := readInput(filename)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(b, c)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing yaml from %s\n", *filename)
		os.Exit(exitInvalidFile)
	}

	return c, err
}

func mustAddResolverFromDataOrDie(origin string, rc dynp.ResolverChain, data []byte) dynp.ResolverChain {
	// determine type
	var f interface{}
	err := json.Unmarshal(data, &f)
	if err != nil {
		err = yaml.Unmarshal(data, &f)
		if err != nil {
			// this is neither json nor yaml
			fmt.Fprintf(os.Stderr, "--param-file %s must be JSON or YAML\n", origin)
			os.Exit(exitInvalidFile)
		}
		switch f.(type) {
		case map[interface{}]interface{}:
			// ok
			r, err := dynp.NewYAMLResolverFromString(string(data))
			if err != nil {
				fmt.Fprintf(os.Stderr, "unable to resolve params from yaml: %s\n", err)
				os.Exit(exitFileErr)
			}
			rc = append(rc, r)
		default:
			fmt.Fprintf(os.Stderr, "--param-file %s must be JSON or YAML\n", origin)
			os.Exit(exitInvalidFile)
		}

	} else {
		r, err := dynp.NewJSONResolverFromString(string(data))
		if err != nil {
			fmt.Fprintf(os.Stderr, "unable to resolve params from json: %s\n", err)
			os.Exit(exitFileErr)
		}
		rc = append(rc, r)

	}

	return rc
}

func resolveParams(ipvsconfig *integration.IPVSConfig) (*integration.IPVSConfig, error) {

	cfg := config.Config()

	intfAddrMap := make(map[string]string)
	if cfg.ParamsHostNetwork == true {
		intfs, err := net.Interfaces()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Specified dynamic parameter from local network interfaces, but unable to query them: %s\n", err)
			os.Exit(exitNetErr)
		}
		for _, intf := range intfs {
			addrs, err := intf.Addrs()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Specified dynamic parameter from local network interfaces, but unable to query details: %s\n", err)
				os.Exit(exitNetErr)
			}
			for idx, addr := range addrs {
				value := addr.String()
				i := strings.LastIndexByte(value, '/')
				if i > 0 {
					value = value[0:i]
				}
				key := fmt.Sprintf("host.%s", intf.Name)
				if idx == 0 {
					intfAddrMap[key] = value
				}
				key = fmt.Sprintf("host.%s_%d", intf.Name, idx)
				intfAddrMap[key] = value
			}
		}
	}

	envMap := make(map[string]string)
	if cfg.ParamsHostEnv == true {
		for _, e := range os.Environ() {
			a := strings.Split(e, "=")
			if len(a) == 2 {
				envMap[fmt.Sprintf("env.%s", a[0])] = a[1]
			}
		}
	}

	// set up resolvers
	mrHostNetwork := dynp.NewMapResolver().
		With(intfAddrMap).
		With(envMap)

	rc := dynp.ResolverChain{mrHostNetwork}

	if len(cfg.ParamsFiles) > 0 {
		for _, pf := range cfg.ParamsFiles {
			if len(pf) == 0 {
				continue
			}
			data, err := ioutil.ReadFile(pf)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Unable to read from parameter file: %s", err)
				os.Exit(exitFileErr)
			}

			rc = mustAddResolverFromDataOrDie(pf, rc, data)
		}
	}
	if len(cfg.ParamsURLs) > 0 {
		for _, url := range cfg.ParamsURLs {
			if len(url) == 0 {
				continue
			}
			resp, err := http.Get(url)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Unable to fetch parameters from url: %s", err)
				os.Exit(exitNetErr)
			}
			defer resp.Body.Close()

			data, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Unable to fetch parameters from url: %s", err)
				os.Exit(exitNetErr)
			}

			rc = mustAddResolverFromDataOrDie(url, rc, data)
		}
	}

	// forward to model using resolvers
	res, err := ipvsconfig.ResolveParams(rc)

	return res, err
}

// MustGetCurrentConfig queries the current IPVS configuration
// or exits in case of an error.
func MustGetCurrentConfig() *integration.IPVSConfig {
	l := config.Config().Logger()
	// retrieve current config
	currentConfig := integration.NewIPVSConfigWithLogger(l)
	err := currentConfig.Get()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to get current ipvs config: %s", err)

		if _, ok := err.(*integration.IPVSHandleError); ok {
			os.Exit(exitIpvsErrHandle)
		}
		if _, ok := err.(*integration.IPVSQueryError); ok {
			os.Exit(exitIpvsErrQuery)
		}
		os.Exit(exitUnknown)
	}

	return currentConfig
}
