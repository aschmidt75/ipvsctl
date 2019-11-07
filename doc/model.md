# ipvsctl - User Documentation

## Model

ipvsctl allows for the model to be expressed in yaml format. A model can be applied or validated through a file or
via STDIN. If no `-f` (file) parameter is given, it reads from `/etc/ipvsctl.yaml`.

### Model elements

#### Addresses

Addresses are representated in combined formats of protocol, ip address and port: `[PROTO]://<IP>:[PORT]`.
Protocol may be `tcp`, `udp` and `sctp`. In Service addresses, the procotol part is mandatory. In Destination addresses it
must be omitted since the protocol of destinations are equal to the service.

IP address part is mandatory. Currently, only IPv4 addresses are supported.

Port is mandatory for services and optional for destinations. If it is omitted in destionations, the port number of
the service is used.

Scheduler names are the valid ipvsadm scheduler names (rr, wrr, lc, wlc, lblc, lblcr, dh, sh, sed, nq). For more details
on schedulers, please see the manpage of ipvsadm.

Valid Forwarder names are `nat` for NAT/Masquerading, `tunnel` for IPIP Tunneling and `direct` for direct routing/gatewaying.

#### Services

Top-Level element `services` is an array of service items. A service contains a mandatory `address`, an optional
scheduler `sched` and optional `destination` items (0..N), e.g.:

```yaml
services:
    - address: tcp://10.0.0.1:8080
      sched: wrr
      destinations:
      - (...)
    - address: (...)
```

#### Destinations

`destination` elements may appear under `services`. A destination is composed of an `address`, an optional `weight` and an optional
`forward` item. 

If no weight is given, `0` is assumed. This behaviour is different from ipvsadm. If no forward is given, the default `direct` is assumed.
Please check ipvsadm's manpage for details.

The address may not contain a protocol, since it is identical to that of the services. It must contain an IP address, only IPv4 is
currently supported. It may contain a port.

```yaml
      destinations:
      - address: 192.168.10.10:80
        forward: nat
        weight: 300
      - address: (...)
    
```

#### Defaults

Users may specify model-wide default values for

* Ports
* Weights
* Forwards
* Schedulers

Whenever a model element misses a part (e.g. a weight), ipvsctl tries to take it from the top-level `defaults` sections. 

```yaml
defaults:
    port: 8008
    weight: 100
    sched: wrr
    forward: nat
```

All items in `defaults` are optional.
