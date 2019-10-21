# ipvsctl

`ipvsctl` is a command line frontend for IPVS using desired state configuration. It reads an IPVS services/destinations model from a YAML file, detects changes to the current configuration and applies those changes. It can pull parameters dynamically from the environment, files or URLs.

It is meant as an add-on to ipvsadm, where changes can be applied from models instead of ad-hoc commands.

## Features

* Adding, Updating and Deleting services and destinations using YAML models
* Services using TCP,UDP,SCTP and FWMARK
* All schedulers, all forwards
* Setting Weights on destinations
* Setting addresses from dynamic parameters (e.g. from environment, files, uris.)

Currently not supported

* IPv6 addresses are not yet supported
* Timeouts, Netmasks, Scheduling flags, Statistics, Thresholds are not supported yet

## Example

```
# cat >/tmp/ipvsconf <<EOF
services:
- address: tcp://${host.eth0}:80
  sched: rr
  destinations:
  - address: 10.50.0.1:${env.MYPORT}
    forward: nat
  - address: 10.50.0.2:${env.MYPORT}
    forward: nat
EOF

# MYPORT=8080 ipvsctl apply -f /tmp/ipvsconf

# ipvsctl get
services:
- address: tcp://10.1.2.3:80
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
TCP  10.1.2.3:80 rr
  -> 10.50.0.1:8080               Masq    0      0          0
  -> 10.50.0.2:8080               Masq    0      0          0    

# ipvsctl -v set weight 100 --service tcp://10.1.2.3:80 --destination 10.50.0.1:8080
INFO Updated weight to 100 for service tcp://10.1.2.3:80/10.50.0.1:8080

# ipvsadm -Ln
[...]
  -> 10.50.0.1:8080               Masq    100    0          0
  -> 10.50.0.2:8080               Masq    0      0          0
```

## Prerequisites

* go 1.13
* Linux
* ipvs kernel modules installed and loaded

## Build

This project builds correctly for Linux only.

```bash
$ CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -v -o release/ipvsctl ipvsctl.go
```

## Test

ipvsctl contains two kinds of tests: unit tests for only small portions of the code and
end-to-end tests for all functions.

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
76 tests, 0 failures
```

## License

(C) 2019 @aschmidt75, Apache 2.0 license
except package ipvs, integrated from https://github.com/docker/libnetwork (C) 2015 Docker, Inc. Apache 2.0 license