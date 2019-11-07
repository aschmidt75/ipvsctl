# ipvsctl - User Documentation

## Dynamic parameters

ipvsctl's model may contain dynamic parameters. Placeholders such as ${myvar} are read from a source other than the
model file. Currently, valid source are OS environment, YAML and JSON files, and YAML and JSON data pulled via HTTP.
Not all model elements may contain parameters. At present, only addresses (both in services and destinations) may contain
parameters, which are resolved at runtime. A model entry is processed as a string and may contain multiple dynamic
parameters. Parameters can embed other parameters (see examples below)

Command line switches to ipvsctl determine, from what sources parameters are pulled and resolved:

```
      --params-network   Dynamic parameters. Add every network interface name as resolvable ip address, e.g. net.eth0
      --params-env       Dynamic parameters. Add every environment entry, e.g. env.port=<ENV VAR "port">
      --params-file      Dynamic parameters. Add parameters from yaml or json file. (default [""])
      --params-url       Dynamic parameters. Add parameters from yaml or json resource given by URL. (default [""])
```

These switches become effective in conjunction with `apply` or `validate` commands. `--params-file` and `--params-url` may
be specified multiple times, and are processed in that order (first entry found wins).

Parameter substitution uses `${` as the beginning marker and `}` as the end marker. So `${my.var}` is valid whereas
`$other.{var}` is not.

If the model contains a parameter which cannot by resolved by the sources given, the validation fails and ipvsctl terminates.

### Parameters from Environment

When using `--params-env`, ipvsctl allows access to all environment entries, prepended by the `env.` prefix. E.g. 

```bash
# MYPORT=8080 ADDR=1.2.3.4 ipvsctl --params-env apply 
```

allows for the model file to contain an address element such as

```
    - address: tcp://${ADDR}:${MYPORT}
```

### Parameters from local network

When using `--params-network`, ipvsctl allows access to the IP addresses of all network interfaces on the host, prepended
by `host.`, e.g.:

```bash
# ipvsctl --params-network apply 
```

allows for the model file to contain an address element such as

```
    - address: tcp://${host.eth0}:80
```

In case a network adapter has more than one IP address, ipvsctl allows access to all of them by adding a zero-based index, e.g.
`- address: tcp://${host.eth0_1}:80` (eth0_1) resolves to the second address of the interface.


### Parameters from files

When using `--params-file`, ipvsctl reads the given file, which must be JSON or YAML. All entries are made available, e.g.:

```bash
# cat p.yaml
ports:
  web: 8443
  api: 9443

# ipvsctl --params-files=p.yaml apply 
```

allows for the model file to contain an address element such as

```
    - address: tcp://10.0.0.1:${ports.web}
```

`--params-file` may be specified multiple times.

### Parameters from URLs

`--params-url` works similar to `--params-file`, but HTTP-GETs the YAML or JSON content from a url, e.g.:

```bash
# ipvsctl --params-url=http://my.config.local:8080/ apply 
```

### Nested parameters

Parameters may be nested, e.g.:

```bash
# cat p.yaml
backends:
  dev: localhost
  test: 192.168.1.10
  stage: 10.0.0.1

# CURENV=test ipvsctl --params-env --params-files=p.yaml apply 
```

allows for

```
    - address: tcp://${backends.${CURENV}}:8080
```

`CURENV` resolves to `test` (from env), and thus `backends.test` resolves to `192.168.1.10` (from p.yaml), so the resulting line is

```
    - address: tcp://192.168.1.10:8080
```

This works in conjunction with all sources.

