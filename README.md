# ipvsctl

`ipvsctl` is a command line frontend for IPVS using desired state configuration. It reads an IPVS services/destinations model from a YAML file, detects changes to the current configuration and applies the changes. 
It is meant as an add-on ipvsadm, where changes can be applied from models instead of ad-hoc commands.

## Example

```
# cat >/tmp/ipvsconf <<EOF
services:
- address: tcp://10.1.2.3:80
  sched: rr
  destinations:
  - address: 10.50.0.1:8080
    forward: nat
  - address: 10.50.0.2:8080
    forward: nat
EOF

# ipvsctl apply -f /tmp/ipvsconf
```

## Prerequisites

* go 1.12
* Linux
* ipvs kernel module installed

## Build

## License

(C) 2019 @aschmidt75, Apache 2.0 license
except package ipvs, integrated from https://github.com/docker/libnetwork (C) 2015 Docker, Inc. Apache 2.0 