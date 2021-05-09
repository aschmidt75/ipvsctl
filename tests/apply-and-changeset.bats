#!/usr/bin/env bats

IPVSCTL=$(which ipvsctl)
if [ -z "${IPVSCTL}" ]; then
	IPVSCTL="$(dirname $BATS_TEST_FILENAME)/../release/ipvsctl"
fi


@test "given any of the model files applied in sequence, when i build a changeset for the same model, it must always be empty" {
    ipvsadm -C

    for fx in $(ls -1 fixtures/*.yaml | grep -v invalid | grep -v params-); do
        run $IPVSCTL apply -f $fx
        [ "$status" -eq 0 ]

        run $IPVSCTL changeset -f $fx

        [ "$status" -eq 0 ]
        [[ "${output}" =~ ^\{\}$ ]] 

    done
    ipvsadm -C
}

