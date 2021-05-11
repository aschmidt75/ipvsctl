#!/usr/bin/env bats

IPVSCTL="$(dirname $BATS_TEST_FILENAME)/../dist/ipvsctl"
if [ ! -x "${IPVSCTL}" ]; then   
    IPVSCTL=$(which ipvsctl)
    if [ ! -x "${IPVSCTL}" ]; then   
        echo ERROR unable to find ipvsctl in local dist or in path
        exit 1
    fi
fi

@test "given a clean ipvs config, when i build a changeset for an empty file, it should be empty" {
    ipvsadm -C

	run $IPVSCTL changeset -f fixtures/apply-empty.yaml

	[ "$status" -eq 0 ]
    [[ "${output}" =~ ^\{\}$ ]]

    ipvsadm -C
}

@test "given a clean ipvs config, when i build a changeset for changeset-service-1, it should contain only that service" {
    ipvsadm -C

	run $IPVSCTL changeset -f fixtures/changeset-service-1.yaml

	[ "$status" -eq 0 ]
    [[ "${output}" =~ add-service ]]
    [[ ! "${output}" =~ delete-service ]]
    [[ ! "${output}" =~ update-service ]]
    [[ ! "${output}" =~ add-destination ]]
    [[ ! "${output}" =~ delete-destination ]]
    [[ ! "${output}" =~ update-destination ]]

    [[ "${output}" =~ address:[[:blank:]]tcp://1\.2\.3\.4:80 ]]
    [[ "${output}" =~ sched:[[:blank:]]rr ]]

    ipvsadm -C
    
}

@test "given ipvs config with changeset-service-1 applied, when i build a changeset for the same file, it should be empty" {
    ipvsadm -C

    $IPVSCTL apply -f fixtures/changeset-service-1.yaml
	run $IPVSCTL changeset -f fixtures/changeset-service-1.yaml

	[ "$status" -eq 0 ]
    [[ "${output}" =~ ^\{\}$ ]]

    ipvsadm -C
    
}

@test "given ipvs config with changeset-service-1 applied, when i build a changeset for changeset-service-2, it should contain an add-service item only" {
    ipvsadm -C

    $IPVSCTL apply -f fixtures/changeset-service-1.yaml
	run $IPVSCTL changeset -f fixtures/changeset-service-2.yaml

	[ "$status" -eq 0 ]
    [[ "${output}" =~ add-service ]]
    [[ ! "${output}" =~ delete-service ]]
    [[ ! "${output}" =~ update-service ]]
    [[ ! "${output}" =~ add-destination ]]
    [[ ! "${output}" =~ delete-destination ]]
    [[ ! "${output}" =~ update-destination ]]

    [[ "${output}" =~ address:[[:blank:]]tcp://1\.2\.3\.5:80 ]]
    [[ "${output}" =~ sched:[[:blank:]]wrr ]]

    ipvsadm -C
    
}

@test "given ipvs config with changeset-service-2 applied, when i build a changeset for changeset-service-3, it should contain an delete-service item only" {
    ipvsadm -C

    $IPVSCTL apply -f fixtures/changeset-service-2.yaml
	run $IPVSCTL changeset -f fixtures/changeset-service-3.yaml

	[ "$status" -eq 0 ]
    [[ ! "${output}" =~ add-service ]]
    [[ "${output}" =~ delete-service ]]
    [[ ! "${output}" =~ update-service ]]
    [[ ! "${output}" =~ add-destination ]]
    [[ ! "${output}" =~ delete-destination ]]
    [[ ! "${output}" =~ update-destination ]]

    [[ "${output}" =~ address:[[:blank:]]tcp://1\.2\.3\.5:80 ]]

    ipvsadm -C
    
}

@test "given ipvs config with changeset-service-3 applied, when i build a changeset for changeset-service-4, it should contain an update-service item only" {
    ipvsadm -C

    $IPVSCTL apply -f fixtures/changeset-service-3.yaml
	run $IPVSCTL changeset -f fixtures/changeset-service-4.yaml

	[ "$status" -eq 0 ]
    [[ ! "${output}" =~ add-service ]]
    [[ ! "${output}" =~ delete-service ]]
    [[ "${output}" =~ update-service ]]
    [[ ! "${output}" =~ add-destination ]]
    [[ ! "${output}" =~ delete-destination ]]
    [[ ! "${output}" =~ update-destination ]]

    [[ "${output}" =~ address:[[:blank:]]tcp://1\.2\.3\.4:80 ]]
    [[ "${output}" =~ sched:[[:blank:]]wrr ]]

    ipvsadm -C
    
}

@test "given a clean ipvs config, when i build a changeset for changeset-destination-1, it should contain only that service with a single destination" {
    ipvsadm -C

	run $IPVSCTL changeset -f fixtures/changeset-destination-1.yaml

	[ "$status" -eq 0 ]
    [[ "${output}" =~ add-service ]]
    [[ ! "${output}" =~ delete-service ]]
    [[ ! "${output}" =~ update-service ]]
    [[ ! "${output}" =~ add-destination ]]
    [[ ! "${output}" =~ delete-destination ]]
    [[ ! "${output}" =~ update-destination ]]

    [[ "${output}" =~ address:[[:blank:]]tcp://1\.2\.3\.4:80 ]]
    [[ "${output}" =~ sched:[[:blank:]]rr ]]
    [[ "${output}" =~ address:[[:blank:]]10\.0\.0\.1 ]]
    [[ "${output}" =~ forward:[[:blank:]]direct ]]
    [[ "${output}" =~ weight:[[:blank:]]100 ]]

    ipvsadm -C
    
}

@test "given ipvs config with changeset-destination-1 applied, when i build a changeset for the same file, it should be empty" {
    ipvsadm -C

    $IPVSCTL apply -f fixtures/changeset-destination-1.yaml
	run $IPVSCTL changeset -f fixtures/changeset-destination-1.yaml

	[ "$status" -eq 0 ]
    [[ "${output}" =~ ^\{\}$ ]]

    ipvsadm -C
    
}

@test "given ipvs config with changeset-destination-1 applied, when i build a changeset for changeset-destination-2, it should contain an add-destination item only" {
    ipvsadm -C

    $IPVSCTL apply -f fixtures/changeset-destination-1.yaml
	run $IPVSCTL changeset -f fixtures/changeset-destination-2.yaml

	[ "$status" -eq 0 ]
    [[ ! "${output}" =~ add-service ]]
    [[ ! "${output}" =~ delete-service ]]
    [[ ! "${output}" =~ update-service ]]
    [[ "${output}" =~ add-destination ]]
    [[ ! "${output}" =~ delete-destination ]]
    [[ ! "${output}" =~ update-destination ]]

    [[ "${output}" =~ address:[[:blank:]]10\.0\.0\.2 ]]
    [[ "${output}" =~ forward:[[:blank:]]direct ]]
    [[ "${output}" =~ weight:[[:blank:]]200 ]]

    ipvsadm -C
    
}

@test "given ipvs config with changeset-destination-1 applied, when i build a changeset for changeset-destination-3, it should contain an delete-destination item only" {
    ipvsadm -C

    $IPVSCTL apply -f fixtures/changeset-destination-1.yaml
	run $IPVSCTL changeset -f fixtures/changeset-destination-3.yaml

	[ "$status" -eq 0 ]
    [[ ! "${output}" =~ add-service ]]
    [[ ! "${output}" =~ delete-service ]]
    [[ ! "${output}" =~ update-service ]]
    [[ ! "${output}" =~ add-destination ]]
    [[ "${output}" =~ delete-destination ]]
    [[ ! "${output}" =~ update-destination ]]

    [[ "${output}" =~ address:[[:blank:]]10\.0\.0\.1 ]]

    ipvsadm -C
    
}

@test "given ipvs config with changeset-destination-2 applied, when i build a changeset for changeset-destination-4, it should contain an update-destination item only" {
    ipvsadm -C

    $IPVSCTL apply -f fixtures/changeset-destination-2.yaml
	run $IPVSCTL changeset -f fixtures/changeset-destination-4.yaml

	[ "$status" -eq 0 ]
    [[ ! "${output}" =~ add-service ]]
    [[ ! "${output}" =~ delete-service ]]
    [[ ! "${output}" =~ update-service ]]
    [[ ! "${output}" =~ add-destination ]]
    [[ ! "${output}" =~ delete-destination ]]
    [[ "${output}" =~ update-destination ]]

    [[ "${output}" =~ address:[[:blank:]]10\.0\.0\.1 ]]
    [[ "${output}" =~ forward:[[:blank:]]tunnel ]]
    [[ "${output}" =~ weight:[[:blank:]]200 ]]

    [[ ! "${output}" =~ address:[[:blank:]]10\.0\.0\.2 ]]

    ipvsadm -C
    
}

@test "given any of the model files applied, when i build a changeset for the same model, it must always be empty" {

    for fx in $(ls -1 fixtures/*.yaml | grep -v invalid | grep -v params-); do
        ipvsadm -C
        $IPVSCTL apply -f $fx
        if [[ $? -eq 0 ]]; then
            run $IPVSCTL changeset -f $fx

            [ "$status" -eq 0 ]
            [[ "${output}" =~ ^\{\}$ ]] 
        fi
    done
    ipvsadm -C
}

