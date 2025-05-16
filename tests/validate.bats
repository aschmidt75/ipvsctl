#!/usr/bin/env bats

IPVSCTL="$(dirname $BATS_TEST_FILENAME)/../dist/ipvsctl"
if [ ! -x "${IPVSCTL}" ]; then   
    IPVSCTL=$(which ipvsctl)
    if [ ! -x "${IPVSCTL}" ]; then   
        echo ERROR unable to find ipvsctl in local dist or in path
        exit 1
    fi
fi

@test "given a correct model, when i validate it, it should pass" {
	run $IPVSCTL validate -f fixtures/apply-single-service.yaml

	[ "$status" -eq 0 ]

	run $IPVSCTL validate -f fixtures/apply-single-service-single-destination.yaml

	[ "$status" -eq 0 ]

	run $IPVSCTL validate -f fixtures/apply-single-service-multiple-destinations.yaml

	[ "$status" -eq 0 ]

	run $IPVSCTL validate -f fixtures/apply-two-services.yaml

	[ "$status" -eq 0 ]

	run $IPVSCTL validate -f fixtures/apply-multiple-services.yaml

	[ "$status" -eq 0 ]
}

@test "given incorrect models, when i validate them, it should fail (reason: bad service addresses)" {
	TMPF=$(mktemp)

	echo -e "services:\n - address: no-such-proto:/123\n" >$TMPF
	run $IPVSCTL validate -f $TMPF

	[[ "$status" -ne 0 ]]

	echo -e "services:\n - address: tcp:/1.2.3.4\n" >$TMPF
	run $IPVSCTL validate -f $TMPF

	[[ "$status" -ne 0 ]]

	echo -e "services:\n - address: tcp://999.999.nos.uch.ipa.ddr/\n" >$TMPF
	run $IPVSCTL validate -f $TMPF

	[[ "$status" -ne 0 ]]

	echo -e "services:\n - address: fwmark:NaN\n" >$TMPF
	run $IPVSCTL validate -f $TMPF

	[[ "$status" -ne 0 ]]

	echo -e "services:\n - address: fwmark:89647\n" >$TMPF
	run $IPVSCTL validate -f $TMPF

	[[ "$status" -ne 0 ]]

    # ipv6 not supported yet
	echo -e "services:\n - address: udp://[fe80:::1]:89/\n" >$TMPF
	run $IPVSCTL validate -f $TMPF

	[[ "$status" -ne 0 ]]

	rm $TMPF

}

@test "given incorrect models, when i validate them, it should fail (reason: bad scheduler names)" {
	TMPF=$(mktemp)

	echo -e "services:\n - address: tcp://1.2.3.4/\n   sched: xrr\n" >$TMPF
	run $IPVSCTL validate -f $TMPF

	[[ "$status" -ne 0 ]]

	rm $TMPF
}

@test "given incorrect models, when i validate them, it should fail (reason: bad destination addresses)" {
	TMPF=$(mktemp)

	echo -e "services:\n  - address: tcp://1.2.3.4:80\n    destinations:\n    - address: \"\"\n" >$TMPF
	run $IPVSCTL validate -f $TMPF

	[[ "$status" -ne 0 ]]

	echo -e "services:\n  - address: tcp://1.2.3.4:80\n    destinations:\n    - address: no.such.ip\n" >$TMPF
	run $IPVSCTL validate -f $TMPF

	[[ "$status" -ne 0 ]]

	echo -e "services:\n  - address: tcp://1.2.3.4:80\n    destinations:\n    - address: 10.0.0.1:BADPORT\n" >$TMPF
	run $IPVSCTL validate -f $TMPF

	[[ "$status" -ne 0 ]]

	echo -e "services:\n  - address: tcp://1.2.3.4:80\n    destinations:\n    - address: 10.0.0.1:89647\n" >$TMPF
	run $IPVSCTL validate -f $TMPF

	[[ "$status" -ne 0 ]]

	rm $TMPF
}

@test "given incorrect models, when i validate them, it should fail (reason: bad forward)" {
	TMPF=$(mktemp)

	echo -e "services:\n  - address: tcp://1.2.3.4:80\n    destinations:\n    - address: 10.0.0.1:90\n      forward: nsf" >$TMPF
	run $IPVSCTL validate -f $TMPF

	[[ "$status" -ne 0 ]]

	rm $TMPF
}

@test "given incorrect models, when i validate them, it should fail (reason: bad weight)" {
	TMPF=$(mktemp)

	echo -e "services:\n  - address: tcp://1.2.3.4:80\n    destinations:\n    - address: 10.0.0.1:90\n      weight: NaN" >$TMPF
	run $IPVSCTL validate -f $TMPF

	[[ "$status" -ne 0 ]]

	echo -e "services:\n  - address: tcp://1.2.3.4:80\n    destinations:\n    - address: 10.0.0.1:90\n      weight: 89647" >$TMPF
	run $IPVSCTL validate -f $TMPF

	[[ "$status" -ne 0 ]]

	rm $TMPF
}

@test "given incorrect defaults, when i validate them, it should fail (bad port)." {
	TMPF=$(mktemp)

    echo -e "defaults:\n  port: 8374284\n" >$TMPF
	run $IPVSCTL validate -f $TMPF

	[[ "$status" -ne 0 ]]

	rm $TMPF
}

@test "given incorrect defaults, when i validate them, it should fail (bad scheduler)." {
	TMPF=$(mktemp)

    echo -e "defaults:\n  port: 80\n  sched: nosuchsched\n" >$TMPF
	run $IPVSCTL validate -f $TMPF

	[[ "$status" -ne 0 ]]

	rm $TMPF
}

@test "given incorrect defaults, when i validate them, it should fail (bad forward)." {
	TMPF=$(mktemp)

    echo -e "defaults:\n  port: 80\n  sched: wrr\n  forward: blackhole\n" >$TMPF
	run $IPVSCTL validate -f $TMPF

	[[ "$status" -ne 0 ]]

	rm $TMPF
}

@test "given incorrect defaults, when i validate them, it should fail (bad weight)." {
	TMPF=$(mktemp)

    echo -e "defaults:\n  port: 80\n  sched: wrr\n  forward: tunnel\n  weight: 7462943" >$TMPF
	run $IPVSCTL validate -f $TMPF

	[[ "$status" -ne 0 ]]

	rm $TMPF
}

@test "given a model with non-unique service(s), when i validate it, it should fail" {
	TMPF=$(mktemp)

	run $IPVSCTL validate -f fixtures/validate-unique-services-invalid.yaml

	[ "$status" -ne 0 ]

	rm $TMPF
}

@test "given a model with non-unique destination(s), when i validate it, it should fail" {
	TMPF=$(mktemp)

	run $IPVSCTL validate -f fixtures/validate-unique-destinations-invalid.yaml

	[ "$status" -ne 0 ]

	rm $TMPF
}
