package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"gopkg.in/yaml.v2"

	ipvs "github.com/aschmidt75/ipvsctl/ipvs"
	"github.com/aschmidt75/ipvsctl/model"
	cli "github.com/jawher/mow.cli"
	log "github.com/sirupsen/logrus"
)

func protoNumToStr(service *ipvs.Service) string {
	switch service.Protocol {
	case 17:
		return "udp"
	case 6:
		return "tcp"
	case 132:
		return "sctp"
	default:
		return "N/A"
	}
}

// Get implements the "get" cli command
func Get(cmd *cli.Cmd) {
	cmd.Spec = "[-o=<OUTFORMAT>]"
	var (
		outformat = cmd.StringOpt("o output-format", "text", "output format, one of: text. Default: text")
	)

	cmd.Action = func() {
		log.Trace("Querying ipvs data...")

		ipvs, err := ipvs.New("")
		if err != nil {
			log.Error("Unable to create IPVS handle. Is the kernel module installed and active?")
			os.Exit(exitIpvsErrHandle)
		}
		log.Tracef("%#v\n", ipvs)

		res := model.IPVSConfig{}

		services, err := ipvs.GetServices()
		if err != nil {
			log.Error("Unable to query IPVS services. Is the kernel module installed and active?")
			os.Exit(exitIpvsErrQuery)
		}
		log.Tracef("%#v\n", services)
		if services != nil && len(services) > 0 {
			res.Services = make([]*model.Service, len(services))

			for idx, service := range services {
				service, err = ipvs.GetService(service)
				if err != nil {
					log.Error("Unable to query IPVS services. Is the kernel module installed and active?")
					os.Exit(exitIpvsErrQuery)
				}

				log.Tracef("%d -> %#v\n", idx, *service)

				var adrStr = ""
				var fwmark = service.FWMark

				if service.Protocol != 0 {
					protoStr := protoNumToStr(service)
					ipStr := fmt.Sprintf("%s", service.Address)
					var portStr = ""
					if service.Port != 0 {
						portStr = fmt.Sprintf(":%d", service.Port)
					}
					adrStr = fmt.Sprintf("%s://%s%s", protoStr, ipStr, portStr)
				}

				res.Services[idx] = &model.Service{
					Address:   adrStr,
					FWMark:    fwmark,
					SchedName: service.SchedName,
				}

				//
				dests, err := ipvs.GetDestinations(service)
				if err != nil {
					log.Error("Unable to query IPVS destinations. Is the kernel module installed and active?")
					os.Exit(exitIpvsErrQuery)
				}

				if dests != nil && len(dests) > 0 {
					res.Services[idx].Destinations = make([]*model.Destination, len(dests))

					for idxd, dest := range dests {
						log.Tracef("%d -> %#v\n", idxd, *dest)

						var adrStr = fmt.Sprintf("%s:%d", dest.Address, dest.Port)

						res.Services[idx].Destinations[idxd] = &model.Destination{
							Address: adrStr,
							Weight:  dest.Weight,
						}
					}
				}
			}
		}

		switch *outformat {
		case "text":
			for _, service := range res.Services {
				if service.FWMark == 0 {
					fmt.Printf("%s\n", service.Address)
				} else {
					fmt.Printf("fwm=%d\n", service.FWMark)
				}
			}
		case "json":
			b, err := json.Marshal(res)
			if err != nil {
				log.Error("unable to format as json")
				os.Exit(exitErrOutput)
			}
			fmt.Printf("%s", string(b))
		case "yaml":
			b, err := yaml.Marshal(res)
			if err != nil {
				log.Error("unable to format as yaml")
				os.Exit(exitErrOutput)
			}
			fmt.Printf("%s", string(b))
		default:
			log.Error("unsupported output format")
			os.Exit(exitErrOutput)
		}
	}
}
