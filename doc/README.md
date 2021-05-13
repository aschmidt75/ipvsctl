# ipvsctl - User Documentation

## Setting up a playground

- [multipass-playground](playground-multipass.md) shows how to use multipass to set up a vm and run tests etc.

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

## Using ipvsctl programmatically

- [libraryexample1](libraryexample1/) shows how to apply complete models from json.
- [libraryexample2](libraryexample2/) is about working with change sets to modify individual items