#!/usr/bin/env bats

IPVSCTL="$(dirname $BATS_TEST_FILENAME)/../release/ipvsctl"

@test "given a configuration with default, when i apply it, all default port values must have been set correctly." {
	ipvsadm -C

	$IPVSCTL apply -f fixtures/apply-defaults.yaml
	run ipvsadm -L -n

	[ "$status" -eq 0 ]

	[[ "${output}" =~ 1\.2\.3\.4:80 ]]
	[[ "${output}" =~ 5\.6\.7\.8:90 ]]
	[[ "${output}" =~ 9\.10\.11\.12:80 ]]

	[[ "${output}" =~ \-\>[[:blank:]]10\.0\.0\.1:80 ]]
	[[ "${output}" =~ \-\>[[:blank:]]10\.0\.0\.2:80 ]]
	[[ "${output}" =~ \-\>[[:blank:]]10\.0\.0\.3:80 ]]
	[[ "${output}" =~ \-\>[[:blank:]]10\.0\.1\.1:90 ]]
	[[ "${output}" =~ \-\>[[:blank:]]10\.0\.1\.2:90 ]]
	[[ "${output}" =~ \-\>[[:blank:]]10\.0\.2\.1:100 ]]
	[[ "${output}" =~ \-\>[[:blank:]]10\.0\.2\.2:100 ]]

	ipvsadm -C
}

@test "given a configuration with default, when i apply it, all default scheduler values must have been set correctly." {
	ipvsadm -C

	$IPVSCTL apply -f fixtures/apply-defaults.yaml
	run ipvsadm -L -n

	[ "$status" -eq 0 ]

	[[ "${output}" =~ 1\.2\.3\.4[[:blank:]]+wrr ]]
	[[ "${output}" =~ 5\.6\.7\.8[[:blank:]]+wrr ]]
	[[ "${output}" =~ 9\.10\.11\.12[[:blank:]]+rr ]]

	ipvsadm -C
}

@test "given a configuration with default, when i apply it, all default forward values must have been set correctly." {
	ipvsadm -C

	$IPVSCTL apply -f fixtures/apply-defaults.yaml
	run ipvsadm -L -n

	[ "$status" -eq 0 ]

	[[ "${output}" =~ \-\>[[:blank:]]10\.0\.0\.1:80[[:blank:]]+Route[[:blank:]] ]]
	[[ "${output}" =~ \-\>[[:blank:]]10\.0\.0\.1:80[[:blank:]]+Route[[:blank:]] ]]
	[[ "${output}" =~ \-\>[[:blank:]]10\.0\.0\.1:80[[:blank:]]+Route[[:blank:]] ]]
	[[ "${output}" =~ \-\>[[:blank:]]10\.0\.1\.1:90[[:blank:]]+Masq[[:blank:]] ]]
	[[ "${output}" =~ \-\>[[:blank:]]10\.0\.1\.2:90[[:blank:]]+Masq[[:blank:]] ]]
	[[ "${output}" =~ \-\>[[:blank:]]10\.0\.2\.1:100[[:blank:]]+Tunnel[[:blank:]] ]]
	[[ "${output}" =~ \-\>[[:blank:]]10\.0\.2\.2:100[[:blank:]]+Tunnel[[:blank:]] ]]

	ipvsadm -C
}

@test "given a configuration with default, when i apply it, all default weight values must have been set correctly." {
	ipvsadm -C

	$IPVSCTL apply -f fixtures/apply-defaults.yaml
	run ipvsadm -L -n

	[ "$status" -eq 0 ]

	[[ "${output}" =~ \-\>[[:blank:]]10\.0\.0\.1:80[[:blank:]]+Route[[:blank:]]+1 ]]
	[[ "${output}" =~ \-\>[[:blank:]]10\.0\.0\.1:80[[:blank:]]+Route[[:blank:]]+2 ]]
	[[ "${output}" =~ \-\>[[:blank:]]10\.0\.0\.1:80[[:blank:]]+Route[[:blank:]]+3 ]]
	[[ "${output}" =~ \-\>[[:blank:]]10\.0\.1\.1:90[[:blank:]]+Masq[[:blank:]]+2 ]]
	[[ "${output}" =~ \-\>[[:blank:]]10\.0\.1\.2:90[[:blank:]]+Masq[[:blank:]]+2 ]]
	[[ "${output}" =~ \-\>[[:blank:]]10\.0\.2\.1:100[[:blank:]]+Tunnel[[:blank:]]+1 ]]
	[[ "${output}" =~ \-\>[[:blank:]]10\.0\.2\.2:100[[:blank:]]+Tunnel[[:blank:]]+1 ]]

	ipvsadm -C
}
