#!/usr/bin/env bats

IPVSCTL="$(dirname $BATS_TEST_FILENAME)/../release/ipvsctl"

@test "given a model with fixed weight, when i apply a change with --keep-weight, it must keep the original weight" {
    ipvsadm -C

    run $IPVSCTL apply -f fixtures/apply-single-service-single-destination-weight.yaml
    [ "$status" -eq 0 ]

    run $IPVSCTL apply --keep-weights -f fixtures/changeset-destination-weight.yaml
    [ "$status" -eq 0 ]

    run $IPVSCTL get
    [ "$status" -eq 0 ]
	[[ "${output}" =~ weight\:\ 100 ]]

    run $IPVSCTL apply -f fixtures/changeset-destination-weight.yaml
    [ "$status" -eq 0 ]

    run $IPVSCTL get
    [ "$status" -eq 0 ]
	[[ "${output}" =~ weight\:\ 200 ]]

    ipvsadm -C
}
