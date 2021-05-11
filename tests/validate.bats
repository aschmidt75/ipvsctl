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
	echo -e "services:\n - address: no-such-proto:/123\n" >/tmp/ipvsctlbats.yaml
	run $IPVSCTL validate -f /tmp/ipvsctlbats.yaml

	[[ "$status" -ne 0 ]]

	echo -e "services:\n - address: tcp:/1.2.3.4\n" >/tmp/ipvsctlbats.yaml
	run $IPVSCTL validate -f /tmp/ipvsctlbats.yaml

	[[ "$status" -ne 0 ]]

	echo -e "services:\n - address: tcp://999.999.nos.uch.ipa.ddr/\n" >/tmp/ipvsctlbats.yaml
	run $IPVSCTL validate -f /tmp/ipvsctlbats.yaml

	[[ "$status" -ne 0 ]]

	echo -e "services:\n - address: fwmark:NaN\n" >/tmp/ipvsctlbats.yaml
	run $IPVSCTL validate -f /tmp/ipvsctlbats.yaml

	[[ "$status" -ne 0 ]]

	echo -e "services:\n - address: fwmark:89647\n" >/tmp/ipvsctlbats.yaml
	run $IPVSCTL validate -f /tmp/ipvsctlbats.yaml

	[[ "$status" -ne 0 ]]

    # ipv6 not supported yet
	echo -e "services:\n - address: udp://[fe80:::1]:89/\n" >/tmp/ipvsctlbats.yaml
	run $IPVSCTL validate -f /tmp/ipvsctlbats.yaml

	[[ "$status" -ne 0 ]]

}

@test "given incorrect models, when i validate them, it should fail (reason: bad scheduler names)" {
	echo -e "services:\n - address: tcp://1.2.3.4/\n   sched: xrr\n" >/tmp/ipvsctlbats.yaml
	run $IPVSCTL validate -f /tmp/ipvsctlbats.yaml

	[[ "$status" -ne 0 ]]
}

@test "given incorrect models, when i validate them, it should fail (reason: bad destination addresses)" {
	echo -e "services:\n  - address: tcp://1.2.3.4:80\n    destinations:\n    - address: \"\"\n" >/tmp/ipvsctlbats.yaml
	run $IPVSCTL validate -f /tmp/ipvsctlbats.yaml

	[[ "$status" -ne 0 ]]

	echo -e "services:\n  - address: tcp://1.2.3.4:80\n    destinations:\n    - address: no.such.ip\n" >/tmp/ipvsctlbats.yaml
	run $IPVSCTL validate -f /tmp/ipvsctlbats.yaml

	[[ "$status" -ne 0 ]]

	echo -e "services:\n  - address: tcp://1.2.3.4:80\n    destinations:\n    - address: 10.0.0.1:BADPORT\n" >/tmp/ipvsctlbats.yaml
	run $IPVSCTL validate -f /tmp/ipvsctlbats.yaml

	[[ "$status" -ne 0 ]]

	echo -e "services:\n  - address: tcp://1.2.3.4:80\n    destinations:\n    - address: 10.0.0.1:89647\n" >/tmp/ipvsctlbats.yaml
	run $IPVSCTL validate -f /tmp/ipvsctlbats.yaml

	[[ "$status" -ne 0 ]]
}

@test "given incorrect models, when i validate them, it should fail (reason: bad forward)" {
	echo -e "services:\n  - address: tcp://1.2.3.4:80\n    destinations:\n    - address: 10.0.0.1:90\n      forward: nsf" >/tmp/ipvsctlbats.yaml
	run $IPVSCTL validate -f /tmp/ipvsctlbats.yaml

	[[ "$status" -ne 0 ]]
}

@test "given incorrect models, when i validate them, it should fail (reason: bad weight)" {
	echo -e "services:\n  - address: tcp://1.2.3.4:80\n    destinations:\n    - address: 10.0.0.1:90\n      weight: NaN" >/tmp/ipvsctlbats.yaml
	run $IPVSCTL validate -f /tmp/ipvsctlbats.yaml

	[[ "$status" -ne 0 ]]

	echo -e "services:\n  - address: tcp://1.2.3.4:80\n    destinations:\n    - address: 10.0.0.1:90\n      weight: 89647" >/tmp/ipvsctlbats.yaml
	run $IPVSCTL validate -f /tmp/ipvsctlbats.yaml

	[[ "$status" -ne 0 ]]
}

@test "given incorrect defaults, when i validate them, it should fail (bad port)." {
    echo -e "defaults:\n  port: 8374284\n" >/tmp/ipvsctlbats.yaml
	run $IPVSCTL validate -f /tmp/ipvsctlbats.yaml

	[[ "$status" -ne 0 ]]
}

@test "given incorrect defaults, when i validate them, it should fail (bad scheduler)." {
    echo -e "defaults:\n  port: 80\n  sched: nosuchsched\n" >/tmp/ipvsctlbats.yaml
	run $IPVSCTL validate -f /tmp/ipvsctlbats.yaml

	[[ "$status" -ne 0 ]]
}

@test "given incorrect defaults, when i validate them, it should fail (bad forward)." {
    echo -e "defaults:\n  port: 80\n  sched: wrr\n  forward: blackhole\n" >/tmp/ipvsctlbats.yaml
	run $IPVSCTL validate -f /tmp/ipvsctlbats.yaml

	[[ "$status" -ne 0 ]]
}

@test "given incorrect defaults, when i validate them, it should fail (bad weight)." {
    echo -e "defaults:\n  port: 80\n  sched: wrr\n  forward: tunnel\n  weight: 7462943" >/tmp/ipvsctlbats.yaml
	run $IPVSCTL validate -f /tmp/ipvsctlbats.yaml

	[[ "$status" -ne 0 ]]
}

@test "given a model with non-unique service(s), when i validate it, it should fail" {
	run $IPVSCTL validate -f fixtures/validate-unique-services-invalid.yaml

	[ "$status" -ne 0 ]

}

@test "given a model with non-unique destination(s), when i validate it, it should fail" {
	run $IPVSCTL validate -f fixtures/validate-unique-destinations-invalid.yaml

	[ "$status" -ne 0 ]

}
