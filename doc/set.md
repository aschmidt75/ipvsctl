# ipvsctl - User Documentation

## Commands

### set

The `set` command is an ad-hoc style command. It allows for setting specific values of 
a destination, currently the weight. It affects the virtual server tables but not the model files.
It has only effect to weight-based schedulers.

#### CLI spec

```
Usage: ipvsctl set COMMAND [arg...]

change services and destinations

Commands:
  weight       set weight of a single destination
```

and

```
Usage: ipvsctl set weight WEIGHT --service=<SERVICE> --destination=<DESTINATION> [--time=<SECONDS>]

set weight of a single destination

Arguments:
  WEIGHT              Weight [0..65535] (default -1)

Options:
  -s, --service       Handle of service, e.g. tcp://127.0.0.1:80
  -d, --destination   Handle of destination, e.g. 10.0.0.1:80
  -t, --time          Number of seconds, for drain/renew mode (default 0)
```

#### Example: Set weight

Sets the weight of a destination to 100 (immediately):

```bash
# ipvsctl set weight 100 --service=tcp://10.0.0.1:80 --destination=10.2.3.4:8080
```

#### Example: Adjust weight over time

Using `--time` in `set weight` will adjust the current weight of the given service by increments
to the desired value, stretched by a number of seconds indicated by `--time`.

E.g. for a destination that has been drained (weight=0) before, bring  weight back to 100 over
a time frame of 60 seconds. As ipvsctl is usually quiet, use verbose flag to print out status updates.

```bash
# ipvsctl -v set weight 100 --service=tcp://10.0.0.1:80 --destination=10.2.3.4:8080 --time 60
(...)
```