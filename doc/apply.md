# ipvsctl - User Documentation

## Commands

### apply

The purpose of the `apply` command is to take the current active ipvs configuration into a new state describe by a model file. After
a successful call to `apply`, the virtual server table will reflect the entries of the model file.

`apply` reads the input either from a file (default: /etc/ipvsctl.yaml) or STDIN. Input must be YAML-formatted, according to the
specification. 

First, it validates the input according to the desired structure, and the spec of individual fields (e.g. an IP-Adresse or
a scheduler name). It continues only if the input is valid.

It then determines the change set, that is a list of change items which take the current virtual server table into the
new, desired state. Afterweards, the change set is applied item-wise.

#### CLI spec

```
Usage: ipvsctl apply [-f=<FILENAME>] [--keep-weights] [--allowed-actions=<ACTIONS_SPEC>]

apply a new configuration from file or stdin

Options:
  -f                      File to apply. Use - for STDIN (default "/etc/ipvsctl.yaml")
      --keep-weights      Leave weights as they are when updating destinations
      --allowed-actions
                          Comma-separated list of allowed actions.
                          as=Add service, us=update service, ds=delete service,
                          ad=Add destination, ud=update destination, dd=delete destination.
                          Default * for all actions.
                          (default "*")
```

#### Example: Applying from default file

```bash
# cat >/etc/ipvsctl.yaml <<EOF
services:
- address: tcp://10.1.2.3:80
  sched: rr
  destinations:
  - address: 10.50.0.1:8080
    forward: nat
  - address: 10.50.0.2:8080
    forward: nat
EOF

# ipvsctl apply
```

The `apply` command does not have an output as per default. If the command was successful, an exit code of 0 is returned. To see
more details, add `-v` for verbose output or `-d` for debug.

#### Example: Apply from STDIN

Using `-` as a parameter to `-f` makes ipvsctl read the input YAML from STDIN. E.g. to delete all virtual server tables, one
can apply an empty model that contains no services or destinations, like so:

```bash
# echo '{}' | ipvsctl apply -f -
```

#### Example: Keeping existing weights of existing vs tables when applying a model update

The apply command writes all values of a model to the virtual server tables, ignoring e.g. changed weights. To drive
update but keep existing weights (which might have been updated in the mean time using the [set](set.md) command), use
`--keep-weights`:

```bash
# ipvsctl -v apply --keep-weights -f ipvs.yaml
```

#### Example: Limiting actions for certain use cases

The switch `--allowed-actions` limits the kind of actions ipvsctl takes on virtual server table entries. It contains a 
comma-separated list of two-letter tokens, where the first letter can be `a` for add, `u` for update or `d` for delete.
The 2nd letter can be `s` for servies or `d` for destination.

For example, to allow only addition of new items and updating of existing items, one can use `--allowed-actions=as,ad,us,ud`.
This way, ipvsctl would not delete destinations or services:

```bash
# ipvsctl -v apply --allowed-actions=as,ad,us,ud -f ipvs.yaml
```
