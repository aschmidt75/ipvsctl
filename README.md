# ipvsctl

`ipvsctl` is a command line frontend for IPVS using desired state configuration. It reads an IPVS services/destinations model from a YAML file, detects changes to the current configuration and applies those changes. It can pull parameters dynamically from the environment, files or URLs.

It is meant as an add-on to ipvsadm, where changes can be applied from models instead of ad-hoc commands.

[![CircleCI](https://circleci.com/gh/aschmidt75/ipvsctl/tree/master.svg?style=svg)](https://circleci.com/gh/aschmidt75/ipvsctl/tree/master)
[![Go Report Card](https://goreportcard.com/badge/github.com/aschmidt75/ipvsctl)](https://goreportcard.com/report/github.com/aschmidt75/ipvsctl)

## Features

* Adding, Updating and Deleting services and destinations using YAML models
* Services using TCP,UDP,SCTP and FWMARK
* All schedulers, all forwards
* Setting Weights on destinations, keeping existing weights when updating destinations
* Setting addresses from dynamic parameters (e.g. from environment, files, uris.)

Currently not supported

* IPv6 addresses are not yet supported
* Timeouts, Netmasks, Scheduling flags, Statistics, Thresholds are not supported yet

`ipvsctl` is a command line tool, but can also be used as a go library to programmatically work with ipvs in a model based fashion.

## Documentation

Please see [the documentation section in doc/](doc/) for more details on commands, model elements etc. 

## Example

* write a sample ipvs configuration file, using placeholder variables. Define a service with two destinations
* `apply` the configuration, filling placeholders from environment variables and local network interface
* `get` the active ipvs configuration
* use classic `ipvsadm` to view the configuration
* use the `set` command to change the weight of an individual destination

```
# cat >/tmp/ipvsconf <<EOF
services:
- address: tcp://\${host.eth0}:7656
  sched: rr
  destinations:
  - address: 10.50.0.1:\${env.MYPORT}
    forward: nat
  - address: 10.50.0.2:\${env.MYPORT}
    forward: nat
EOF

# MYPORT=8080 ipvsctl --params-network --params-env apply -f /tmp/ipvsconf

# ipvsctl get
services:
- address: tcp://10.1.2.3:7656
  sched: rr
  destinations:
  - address: 10.50.0.2:8080
    forward: nat
  - address: 10.50.0.1:8080
    forward: nat

# ipvsadm -Ln
IP Virtual Server version 1.2.1 (size=4096)
Prot LocalAddress:Port Scheduler Flags
  -> RemoteAddress:Port           Forward Weight ActiveConn InActConn
TCP  10.1.2.3:7656 rr
  -> 10.50.0.1:8080               Masq    0      0          0
  -> 10.50.0.2:8080               Masq    0      0          0    

# ipvsctl -v set weight 100 --service tcp://10.1.2.3:80 --destination 10.50.0.1:8080
INFO Updated weight to 100 for service tcp://10.1.2.3:80/10.50.0.1:8080

# ipvsadm -Ln
[...]
  -> 10.50.0.1:8080               Masq    100    0          0
  -> 10.50.0.2:8080               Masq    0      0          0
```

For using ipvsctl programmatically as a library within your own go code, see [doc/libraryexample1](doc/libraryexample1) and [doc/libraryexample2](doc/libraryexample2).

## Prerequisites

* Linux
* ipvs kernel modules installed and loaded
* for building ipvsctl: go 1.16

## Install

You can build ipvsctl as described below or install one of the versions under the `releases` tab.
`ipvsctl` makes modifications to the ipvs tables, so it either needs to be run as root or equipped
with the appropriate capabilities, e.g.:

```bash
$ VERSION=0.2.3
$ URL=https://github.com/aschmidt75/ipvsctl/releases/download/v${VERSION}/ipvsctl_${VERSION}_$(uname -s)_$(uname -m).tar.gz
$ curl -L $URL | tar xfvz -
$ chmod +x ipvsctl

$ # either run as root ...
$ sudo cp ipvsctl /sbin
$ # .. or open for all users
$ sudo cp ipvsctl /usr/local/bin
$ sudo setcap 'cap_net_admin+eip' /usr/local/bin/ipvsctl 
```

Caution as this allows any user to modify ipvs tables! Please evaluate whether `sudo` or `setcap` is the right approach for you.

## Build

This project builds correctly for Linux only.

```bash
$ make
$ dist/ipvsctl --version
0.2.3
```

## Test

ipvsctl contains three kinds of tests: 
* unit tests for only small portions of the code 
* integration tests (desctructive, will overwrite ipvs tables)
* bats-based end-to-end tests for all cli functions (desctructive, will overwrite ipvs tables)

### Unit and integration tests

Integration tests will touch ipvs tables, so it's required to run as root:

```
# go test -cover ./...
```

### End to end tests

E2E tests run on the linux command line using bats. The test scripts run all commands of
ipvsctl and compare the output against what ipvsadm tells about the underlying ipvs data - or vice versa.

To run the tests, build ipvsctl using the above command, install the necessary prequisites on the host:

* enable ipvs
* install ipvsadm
* install bats
* docker is necessary for testing parameter file pulls from URIs

Test cases can be run like this:

```bash
$ cd tests
$ bats .
 ✓ given any of the model files applied in sequence, when i build a changeset for the same model, it must always be empty
 ✓ given a configuration with defaults, when i apply it, all default port values must have been set correctly.
 ✓ given a configuration with defaults, when i apply it, all default scheduler values must have been set correctly.
[...]
77 tests, 0 failures
```

## License

(C) 2019 @aschmidt75, Apache 2.0 license
except package ipvs, integrated from https://github.com/docker/libnetwork (C) 2015 Docker, Inc. Apache 2.0 license
