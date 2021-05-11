#!/usr/bin/env bats

IPVSCTL="$(dirname $BATS_TEST_FILENAME)/../dist/ipvsctl"
if [ ! -x "${IPVSCTL}" ]; then   
    IPVSCTL=$(which ipvsctl)
    if [ ! -x "${IPVSCTL}" ]; then   
        echo ERROR unable to find ipvsctl in local dist or in path
        exit 1
    fi
fi

# test for a clean ipvs config 
@test "empty ipvs table yields empty yaml" {
	ipvsadm -C
	run $IPVSCTL get 
	[ "$status" -eq 0 ]
	[ "$output" == "{}" ]
}

@test "when i apply an invalid yaml file, it fails" {
	run $IPVSCTL apply -f fixtures/apply-invalidyaml.yaml

	[ "$status" -ne 0 ]
}

@test "when i apply a nonexisting file, it fails" {
	run $IPVSCTL apply -f fixtures/apply-nonexisting.yaml

	[ "$status" -ne 0 ]
}

@test "given ipvs config is clean, when i apply 'empty', it must still be clean" {
	$IPVSCTL apply -f fixtures/apply-empty.yaml
	run ipvsadm -L -n

	[ "$status" -eq 0 ]
	[[ ! "${output}" =~ ":80" ]]

	NL=$(echo "${output}" | wc -l)
	[ "${NL}" = "3" ]

	ipvsadm -C
}

@test "given ipvs config is clean, when i apply single-service, it must appear" {
	$IPVSCTL apply -f fixtures/apply-single-service.yaml
	run ipvsadm -L -n

	[ "$status" -eq 0 ]
	[[ "${output}" =~ 1\.2\.3\.4:80 ]]

	ipvsadm -C
}

@test "given ipvs config with a single service, when i apply 'empty', ipvs config must be clean" {
	ipvsadm -A -t 1.2.3.4:10000 -s wrr
	$IPVSCTL apply -f fixtures/apply-empty.yaml
	run ipvsadm -L -n

	[ "$status" -eq 0 ]
	[[ ! "${output}" =~ 1\.2\.3\.4:10000 ]]

	ipvsadm -C
}

@test "given ipvs config with a service, when i apply 'single-service', ipvs config must contain only the new one" {
	ipvsadm -A -t 5.6.7.8:81 -s wrr
	$IPVSCTL apply -f fixtures/apply-single-service.yaml
	run ipvsadm -L -n

	[ "$status" -eq 0 ]
	[[ "${output}" =~ 1\.2\.3\.4:80 ]]
	[[ ! "${output}" =~ 5\.6\.7\.8:81 ]]

	ipvsadm -C
}

@test "given a clean ipvs config, when i apply 'multiple-services', ipvs config must contain all" {
	ipvsadm -C
	$IPVSCTL apply -f fixtures/apply-multiple-services.yaml
	run ipvsadm -L -n

	[ "$status" -eq 0 ]
	[[ "${output}" =~ TCP[[:blank:]]+10\.99\.0\.1:80[[:blank:]]+rr ]]
	[[ "${output}" =~ TCP[[:blank:]]+10\.99\.0\.2:80[[:blank:]]+wrr ]]
	[[ "${output}" =~ TCP[[:blank:]]+10\.99\.0\.3:80[[:blank:]]+lc ]]
	[[ "${output}" =~ TCP[[:blank:]]+10\.99\.0\.4:80[[:blank:]]+wlc ]]
	[[ "${output}" =~ TCP[[:blank:]]+10\.99\.0\.5:80[[:blank:]]+lblc ]]
	[[ "${output}" =~ TCP[[:blank:]]+10\.99\.0\.6:80[[:blank:]]+lblcr ]]
	[[ "${output}" =~ TCP[[:blank:]]+10\.99\.0\.7:80[[:blank:]]+dh ]]
	[[ "${output}" =~ TCP[[:blank:]]+10\.99\.0\.8:80[[:blank:]]+sh ]]
	[[ "${output}" =~ TCP[[:blank:]]+10\.99\.0\.9:80[[:blank:]]+sed ]]
	[[ "${output}" =~ TCP[[:blank:]]+10\.99\.0\.10:80[[:blank:]]+nq ]]

	[[ "${output}" =~ UDP[[:blank:]]+10\.99\.1\.1:80[[:blank:]]+rr ]]
	[[ "${output}" =~ UDP[[:blank:]]+10\.99\.1\.2:80[[:blank:]]+wrr ]]
	[[ "${output}" =~ UDP[[:blank:]]+10\.99\.1\.3:80[[:blank:]]+lc ]]
	[[ "${output}" =~ UDP[[:blank:]]+10\.99\.1\.4:80[[:blank:]]+wlc ]]
	[[ "${output}" =~ UDP[[:blank:]]+10\.99\.1\.5:80[[:blank:]]+lblc ]]
	[[ "${output}" =~ UDP[[:blank:]]+10\.99\.1\.6:80[[:blank:]]+lblcr ]]
	[[ "${output}" =~ UDP[[:blank:]]+10\.99\.1\.7:80[[:blank:]]+dh ]]
	[[ "${output}" =~ UDP[[:blank:]]+10\.99\.1\.8:80[[:blank:]]+sh ]]
	[[ "${output}" =~ UDP[[:blank:]]+10\.99\.1\.9:80[[:blank:]]+sed ]]
	[[ "${output}" =~ UDP[[:blank:]]+10\.99\.1\.10:80[[:blank:]]+nq ]]

	[[ "${output}" =~ SCTP[[:blank:]]+10\.99\.2\.1:80[[:blank:]]+rr ]]
	[[ "${output}" =~ SCTP[[:blank:]]+10\.99\.2\.2:80[[:blank:]]+wrr ]]
	[[ "${output}" =~ SCTP[[:blank:]]+10\.99\.2\.3:80[[:blank:]]+lc ]]
	[[ "${output}" =~ SCTP[[:blank:]]+10\.99\.2\.4:80[[:blank:]]+wlc ]]
	[[ "${output}" =~ SCTP[[:blank:]]+10\.99\.2\.5:80[[:blank:]]+lblc ]]
	[[ "${output}" =~ SCTP[[:blank:]]+10\.99\.2\.6:80[[:blank:]]+lblcr ]]
	[[ "${output}" =~ SCTP[[:blank:]]+10\.99\.2\.7:80[[:blank:]]+dh ]]
	[[ "${output}" =~ SCTP[[:blank:]]+10\.99\.2\.8:80[[:blank:]]+sh ]]
	[[ "${output}" =~ SCTP[[:blank:]]+10\.99\.2\.9:80[[:blank:]]+sed ]]
	[[ "${output}" =~ SCTP[[:blank:]]+10\.99\.2\.10:80[[:blank:]]+nq ]]

	[[ "${output}" =~ FWM[[:blank:]]+30[[:blank:]]+rr ]]
	[[ "${output}" =~ FWM[[:blank:]]+31[[:blank:]]+wrr ]]
	[[ "${output}" =~ FWM[[:blank:]]+32[[:blank:]]+lc ]]
	[[ "${output}" =~ FWM[[:blank:]]+33[[:blank:]]+wlc ]]
	[[ "${output}" =~ FWM[[:blank:]]+34[[:blank:]]+lblc ]]
	[[ "${output}" =~ FWM[[:blank:]]+35[[:blank:]]+lblcr ]]
	[[ "${output}" =~ FWM[[:blank:]]+36[[:blank:]]+dh ]]
	[[ "${output}" =~ FWM[[:blank:]]+37[[:blank:]]+sh ]]
	[[ "${output}" =~ FWM[[:blank:]]+38[[:blank:]]+sed ]]
	[[ "${output}" =~ FWM[[:blank:]]+39[[:blank:]]+nq ]]

	[[ ! "${output}" =~ 10\.99\.0\.1:80[[:blank:]]+wrr ]]
	[[ ! "${output}" =~ 10\.99\.0\.2:80[[:blank:]]+rr ]]
	[[ ! "${output}" =~ 10\.99\.0\.3:80[[:blank:]]+rr ]]
	[[ ! "${output}" =~ 10\.99\.0\.4:80[[:blank:]]+rr ]]
	[[ ! "${output}" =~ 10\.99\.0\.5:80[[:blank:]]+rr ]]
	[[ ! "${output}" =~ 10\.99\.0\.6:80[[:blank:]]+rr ]]
	[[ ! "${output}" =~ 10\.99\.0\.7:80[[:blank:]]+rr ]]
	[[ ! "${output}" =~ 10\.99\.0\.8:80[[:blank:]]+rr ]]
	[[ ! "${output}" =~ 10\.99\.0\.9.80[[:blank:]]+rr ]]
	[[ ! "${output}" =~ 10\.99\.0\.10:80[[:blank:]]+rr ]]

	[[ ! "${output}" =~ 1\.2\.3\.4:80 ]]
	[[ ! "${output}" =~ 5\.6\.7\.8:81 ]]

	ipvsadm -C
}

@test "given ipvs config with a service, when i apply 'single-service', ipvs config must contain correct scheduler name" {
	ipvsadm -A -t 1.2.3.4:80 -s wrr
	$IPVSCTL apply -f fixtures/apply-single-service.yaml
	run ipvsadm -L -n

	[ "$status" -eq 0 ]
	[[ "${output}" =~ TCP[[:blank:]]+1\.2\.3\.4:80[[:blank:]]+rr ]]
	[[ ! "${output}" =~ TCP[[:blank:]]+1\.2\.3\.4:80[[:blank:]]+wrr ]]

	ipvsadm -C
}

@test "given ipvs config with a service, when i apply 'two-services', ipvs config must contain correct entries (mixed test table)" {
	ipvsadm -A -t 1.2.3.4:80 -s sed				# this must be updated (sed->rr)
	ipvsadm -A -t 10.20.30.40:80 -s wlc			# this must be gone
												# 4.5.6.7.8:81 rr must appear

	$IPVSCTL apply -f fixtures/apply-two-services.yaml
	run ipvsadm -L -n

	[ "$status" -eq 0 ]
	[[ "${output}" =~ TCP[[:blank:]]+1\.2\.3\.4:80[[:blank:]]+rr ]]
	[[ "${output}" =~ UDP[[:blank:]]+5\.6\.7\.8:81[[:blank:]]+lc ]]
	[[ ! "${output}" =~ TCP[[:blank:]]+1\.2\.3\.4:80[[:blank:]]+sed ]]
	[[ ! "${output}" =~ TCP[[:blank:]]+10\.20\.30\.40:80 ]]

	ipvsadm -C
}

