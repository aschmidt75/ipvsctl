#!/usr/bin/env bats

IPVSCTL="$(dirname $BATS_TEST_FILENAME)/../release/ipvsctl"
ASSSD=fixtures/apply-single-service-single-destination.yaml

@test "when i apply with invalid allowed actions, it must fail" {
    run $IPVSCTL apply --allowed-actions=inv -f $ASSSD
    [ "$status" -ne 0 ]
}

@test "given a simple model and a changeset, when i apply with no allowed actions, it must fail" {
    ipvsadm -C

    run $IPVSCTL apply --allowed-actions= -f $ASSSD
    [ "$status" -ne 0 ]

    ipvsadm -C
}

@test "given a simple model and a changeset, when i apply with missing allowed actions, it must fail (no add service allowed)" {
    ipvsadm -C

    $IPVSCTL apply -f $ASSSD
    run $IPVSCTL apply --allowed-actions=us,ds,ad,ud,dd -f fixtures/changeset-destination-5.yaml
    [ "$status" -ne 0 ]

    ipvsadm -C
}

@test "given a simple model and a changeset, when i apply with necessary allowed actions, it must pass (add service allowed)" {
    ipvsadm -C

    $IPVSCTL apply -f $ASSSD
    run $IPVSCTL apply --allowed-actions=as,us,ds,ad,ud,dd -f fixtures/changeset-destination-5.yaml
    [ "$status" -eq 0 ]

    ipvsadm -C
}

@test "given a simple model and a changeset, when i apply with missing allowed actions, it must fail (no update service allowed)" {
    ipvsadm -C

    $IPVSCTL apply -f $ASSSD
    run $IPVSCTL apply --allowed-actions=as,ds,ad,ud,dd -f fixtures/changeset-destination-5.yaml
    [ "$status" -ne 0 ]

    ipvsadm -C
}

@test "given a simple model and a changeset, when i apply with necessary allowed actions, it must pass (update service allowed)" {
    ipvsadm -C

    $IPVSCTL apply -f $ASSSD
    run $IPVSCTL apply --allowed-actions=as,us,ds,ad,ud,dd -f fixtures/changeset-destination-5.yaml
    [ "$status" -eq 0 ]

    ipvsadm -C
}

@test "given a simple model and a changeset, when i apply with missing allowed actions, it must fail (no delete service allowed)" {
    ipvsadm -C

    $IPVSCTL apply -f $ASSSD
    run $IPVSCTL apply --allowed-actions=as,us,ad,ud,dd -f fixtures/changeset-destination-6.yaml
    [ "$status" -ne 0 ]

    ipvsadm -C
}

@test "given a simple model and a changeset, when i apply with necessary allowed actions, it must pass (delete service allowed)" {
    ipvsadm -C

    $IPVSCTL apply -f $ASSSD
    run $IPVSCTL apply --allowed-actions=as,us,ds,ad,ud,dd -f fixtures/changeset-destination-6.yaml
    [ "$status" -eq 0 ]

    ipvsadm -C
}

@test "given a simple model and a changeset, when i apply with missing allowed actions, it must fail (no update destination allowed)" {
    ipvsadm -C

    $IPVSCTL apply -f $ASSSD
    run $IPVSCTL apply --allowed-actions=as,ad,ds,dd -f fixtures/changeset-destination-1.yaml
    [ "$status" -ne 0 ]

    ipvsadm -C
}

@test "given a simple model and a changeset, when i apply with necessary allowed actions, it must pass (update destination allowed)" {
    ipvsadm -C

    $IPVSCTL apply -f $ASSSD
    run $IPVSCTL apply --allowed-actions=us,ud -f fixtures/changeset-destination-1.yaml
    [ "$status" -eq 0 ]

    ipvsadm -C
}

@test "given a simple model and a changeset, when i apply with missing allowed actions, it must fail (no add destination allowed)" {
    ipvsadm -C

    $IPVSCTL apply -f $ASSSD
    run $IPVSCTL apply --allowed-actions=as,us,ds,ud,dd -f fixtures/changeset-destination-2.yaml
    [ "$status" -ne 0 ]

    ipvsadm -C
}

@test "given a simple model and a changeset, when i apply with necessary allowed actions, it must pass (add destination allowed)" {
    ipvsadm -C

    $IPVSCTL apply -f $ASSSD
    run $IPVSCTL apply --allowed-actions=us,ud,ad -f fixtures/changeset-destination-2.yaml
    [ "$status" -eq 0 ]

    ipvsadm -C
}

@test "given a simple model and a changeset, when i apply with missing allowed actions, it must fail (no delete destination allowed)" {
    ipvsadm -C

    $IPVSCTL apply -f $ASSSD
    run $IPVSCTL apply --allowed-actions=as,us,ds,ud -f fixtures/changeset-destination-3.yaml
    [ "$status" -ne 0 ]

    ipvsadm -C
}

@test "given a simple model and a changeset, when i apply with necessary allowed actions, it must pass (delete destination allowed)" {
    ipvsadm -C

    $IPVSCTL apply -f $ASSSD
    run $IPVSCTL apply --allowed-actions=us,dd -f fixtures/changeset-destination-3.yaml
    [ "$status" -eq 0 ]

    ipvsadm -C
}
