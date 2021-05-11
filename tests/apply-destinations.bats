#!/usr/bin/env bats

IPVSCTL="$(dirname $BATS_TEST_FILENAME)/../dist/ipvsctl"
if [ ! -x "${IPVSCTL}" ]; then   
    IPVSCTL=$(which ipvsctl)
    if [ ! -x "${IPVSCTL}" ]; then   
        echo ERROR unable to find ipvsctl in local dist or in path
        exit 1
    fi
fi

@test "given ipvs config is clean, when i apply single-service-single-destination, it must appear correctly" {
	ipvsadm -C

	$IPVSCTL apply -f fixtures/apply-single-service-single-destination.yaml
	run ipvsadm -L -n

	[ "$status" -eq 0 ]
	[[ "${output}" =~ 1\.2\.3\.4:80 ]]
	[[ "${output}" =~ \-\>[[:blank:]]10\.0\.0\.1:80[[:blank:]]+Route[[:blank:]]+0 ]]

	ipvsadm -C
}

@test "given ipvs config is clean, when i apply single-service-multiple-destinations, all destinations must appear correctly" {
	ipvsadm -C

	$IPVSCTL apply -f fixtures/apply-single-service-multiple-destinations.yaml
	run ipvsadm -L -n

	[ "$status" -eq 0 ]
	[[ "${output}" =~ 1\.2\.3\.4:80 ]]
	[[ "${output}" =~ \-\>[[:blank:]]10\.0\.0\.1:80[[:blank:]]+Route[[:blank:]]+1 ]]
	[[ "${output}" =~ \-\>[[:blank:]]10\.0\.0\.2:80[[:blank:]]+Route[[:blank:]]+2 ]]
	[[ "${output}" =~ \-\>[[:blank:]]10\.0\.0\.3:80[[:blank:]]+Route[[:blank:]]+3 ]]
	[[ "${output}" =~ 5\.6\.7\.8:80 ]]
	[[ "${output}" =~ \-\>[[:blank:]]10\.0\.1\.1:90[[:blank:]]+Masq[[:blank:]]+1 ]]
	[[ "${output}" =~ \-\>[[:blank:]]10\.0\.1\.2:90[[:blank:]]+Masq[[:blank:]]+1 ]]
	[[ "${output}" =~ 9\.10\.11\.12:80 ]]
	[[ "${output}" =~ \-\>[[:blank:]]10\.0\.2\.1:100[[:blank:]]+Tunnel[[:blank:]]+1 ]]
	[[ "${output}" =~ \-\>[[:blank:]]10\.0\.2\.2:100[[:blank:]]+Tunnel[[:blank:]]+1 ]]

	ipvsadm -C
}

@test "given ipvs config is equal to single-service-multiple-destinations, when i apply single-service-multiple-destinations-update, all destination updates must appear correctly" {
	ipvsadm -C
	$IPVSCTL apply -f fixtures/apply-single-service-multiple-destinations.yaml
	$IPVSCTL apply -f fixtures/apply-single-service-multiple-destinations-update.yaml

	run ipvsadm -L -n

	[ "$status" -eq 0 ]

	# these must have gone
	[[ ! "${output}" =~ \-\>[[:blank:]]10\.0\.0\.1:80[[:blank:]]+Route ]]
	[[ ! "${output}" =~ \-\>[[:blank:]]10\.0\.1\.1:80[[:blank:]]+Route ]]
	[[ ! "${output}" =~ \-\>[[:blank:]]10\.0\.2\.1:80[[:blank:]]+Route ]]

	# some forwards must have changed
	[[ "${output}" =~ \-\>[[:blank:]]10\.0\.0\.2:80[[:blank:]]+Route ]]
	[[ "${output}" =~ \-\>[[:blank:]]10\.0\.0\.3:80[[:blank:]]+Tunnel ]]
	[[ "${output}" =~ \-\>[[:blank:]]10\.0\.1\.2:90[[:blank:]]+Masq ]]
	[[ "${output}" =~ \-\>[[:blank:]]10\.0\.2\.2:100[[:blank:]]+Route[[:blank:]]+3 ]]

	ipvsadm -C
}