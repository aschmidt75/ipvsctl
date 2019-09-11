#!/usr/bin/env bats

IPVSCTL="$(dirname $BATS_TEST_FILENAME)/../release/ipvsctl"

@test "given a correct model, when i validate it, it should pass" {
	run $IPVSCTL validate -f fixtures/apply-single-service.yaml

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