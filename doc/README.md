# ipvsctl - User Documentation

## Command Reference and Examples

- [get](get.md) is used to retrieve the current active configuration
- [validate](validate.md) validates a mode configuration
- [apply](apply.md) applies a configuration from a model file 
- [changeset](changeset.md) is used to mask the difference between the current active configuration and a model file
- [set](set.md) is used to change settings on individual destinations, e.g. weights

## Model Reference

ipvsctl works on yaml structures, which are described in the [model section](model.md).

## Dynamic parameters

ipvsctl's model may contain [dynamic parameters](dynamicparams.md). Placeholders such as `${myvar}` are read from a source other than the
model file. 