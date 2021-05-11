#!/usr/bin/env bats

IPVSCTL="$(dirname $BATS_TEST_FILENAME)/../dist/ipvsctl"
if [ ! -x "${IPVSCTL}" ]; then   
    IPVSCTL=$(which ipvsctl)
    if [ ! -x "${IPVSCTL}" ]; then   
        echo ERROR unable to find ipvsctl in local dist or in path
        exit 1
    fi
fi

@test "empty ipvs table yields empty yaml" {
	ipvsadm -C
	run $IPVSCTL get 
	[ "$status" -eq 0 ]
	[ "$output" == "{}" ]
}

@test "showing a single service (tcp, rr)" {
	ipvsadm -C
	ipvsadm -A -t 1.2.3.4:10000 -s rr
	run $IPVSCTL get 
	ipvsadm -C

	[ "$status" -eq 0 ]
	[[ "${lines[0]}" == "services:" ]]
	[[ "${lines[1]}" =~ ^-\ address:\ tcp://1\.2\.3\.4:10000$ ]]
	[[ "$output" =~ .*sched:\ rr.* ]]
	[[ ! "$output" =~ destinations ]]
}

@test "showing a single service (tcp, wrr)" {
	ipvsadm -C
	ipvsadm -A -t 1.2.3.4:10000 -s wrr
	run $IPVSCTL get 
	ipvsadm -C

	[ "$status" -eq 0 ]
	[[ "${lines[0]}" == "services:" ]]
	[[ "${lines[1]}" =~ ^-\ address:\ tcp://1\.2\.3\.4:10000$ ]]
	[[ "$output" =~ .*sched:\ wrr.* ]]
	[[ ! "$output" =~ destinations ]]
}

@test "showing a single service (udp, rr)" {
	ipvsadm -C
	ipvsadm -A -u 1.2.3.4:10000 -s rr
	run $IPVSCTL get 
	ipvsadm -C

	[ "$status" -eq 0 ]
	[[ "${lines[0]}" == "services:" ]]
	[[ "${lines[1]}" =~ ^-\ address:\ udp://1\.2\.3\.4:10000$ ]]
	[[ "$output" =~ .*sched:\ rr.* ]]
	[[ ! "$output" =~ destinations ]]
}

@test "showing a single service (fwmark, rr)" {
	ipvsadm -C
	ipvsadm -A -f 746 -s rr
	run $IPVSCTL get 
	ipvsadm -C

	[ "$status" -eq 0 ]
	[[ "${lines[0]}" == "services:" ]]
	[[ "${lines[1]}" =~ ^-\ address:\ fwmark:746$ ]]
	[[ "$output" =~ .*sched:\ rr.* ]]
	[[ ! "$output" =~ destinations ]]
}

@test "showing two services" {
	ipvsadm -C
	ipvsadm -A -t 1.2.3.4:10000 -s rr
	ipvsadm -A -t 5.6.7.8:10001 -s rr
	run $IPVSCTL get 
	ipvsadm -C

	[ "$status" -eq 0 ]
	[[ "${lines[0]}" == "services:" ]]
	[[ "$output" =~ address:\ tcp://1\.2\.3\.4:10000 ]]
	[[ "$output" =~ address:\ tcp://5\.6\.7\.8:10001 ]]
	[[ ! "$output" =~ address:\ tcp://5\.6\.7\.8:10000 ]]
	[[ ! "$output" =~ address:\ udp://5\.6\.7\.8:10001 ]]
	[[ ! "$output" =~ destinations ]]
}

# destinations

@test "showing a single service with a single destination (tcp,rr,masq,def.weight)" {
	ipvsadm -C
	ipvsadm -A -t 1.2.3.4:10000 -s rr
	ipvsadm -a -t 1.2.3.4:10000 -r 10.0.0.1:20000 -m
	run $IPVSCTL get
	ipvsadm -C

	[ "$status" -eq 0 ]
	[[ "${lines[0]}" == "services:" ]]
	[[ "${lines[1]}" =~ ^-\ address:\ tcp://1\.2\.3\.4:10000$ ]]
	[[ "$output" =~ .*sched:\ rr.* ]]
	[[ "$output" =~ destinations ]]
	[[ "$output" =~ \ \ -\ address:\ 10\.0\.0\.1:20000 ]]
	[[ ! "$output" =~ \ \ -\ address:\ 10\.0\.0\.2:20000 ]]
	[[ "$output" =~ .*weight:\ 1.* ]]
	# todo: masq
}

@test "showing a single service with two destinations (tcp,rr,masq,def.weight)" {
	ipvsadm -C
	ipvsadm -A -t 1.2.3.4:10000 -s rr
	ipvsadm -a -t 1.2.3.4:10000 -r 10\.0\.0\.1:20000 -m
	ipvsadm -a -t 1.2.3.4:10000 -r 10\.0\.0\.2:20000 -m
	run $IPVSCTL get
	ipvsadm -C

	[ "$status" -eq 0 ]
	[[ "${lines[0]}" == "services:" ]]
	[[ "${lines[1]}" =~ ^-\ address:\ tcp://1.2.3.4:10000$ ]]
	[[ "$output" =~ .*sched:\ rr.* ]]
	[[ "$output" =~ destinations ]]
	[[ "$output" =~ \ \ -\ address:\ 10\.0\.0\.1:20000 ]]
	[[ "$output" =~ \ \ -\ address:\ 10\.0\.0\.2:20000 ]]
	[[ "$output" =~ .*weight:\ 1.* ]]
	# todo: masq
}