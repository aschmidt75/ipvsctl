# ipvsctl - User Documentation

## Commands

### validate

The `validate` command takes a (yaml-valid) input model and checks its structure and parameters. It reads the model file, 
resolves dynamic parameters and makes sure everything is well-formed so it can be applied successfully (e.g. IP addresses are correct,
scheduler names are ok, weights are valid, etc.)
Validation does not alter anything and can be run as non-root as well.

#### CLI spec

```
Usage: ipvsctl validate [-f=<FILENAME>]

validate a configuration from file or stdin

Options:
  -f           File to apply. Use - for STDIN (default "/etc/ipvsctl.yaml")
```

#### Example

```bash
$ ipvsctl validate -f bad.yaml
ERRO Configuration not valid: unable to parse address (no.such.ip:80). Must be of format <proto>://<host>[:port] or fwmark:<id>.
```

In case of successful validation, exit code is 0 and no output is returned, unless verbose flag is specified:

```bash
$ ipvsctl -v validate -f tests/fixtures/apply-single-service-single-destination.yaml
INFO Configuration valid.
```
