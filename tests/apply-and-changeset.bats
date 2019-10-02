#!/usr/bin/env bats

IPVSCTL="$(dirname $BATS_TEST_FILENAME)/../release/ipvsctl"


@test "given any of the model files applied in sequence, when i build a changeset for the same model, it must always be empty" {
    ipvsadm -C

    for fx in $(ls -1 fixtures/*.yaml); do
        if [[ ! "${fx}" =~ invalid ]]; then
            run $IPVSCTL apply -f $fx
            [ "$status" -eq 0 ]

            run $IPVSCTL changeset -f $fx

            [ "$status" -eq 0 ]
            [[ "${output}" =~ ^\{\}$ ]] 

        fi
    done
    ipvsadm -C
}

