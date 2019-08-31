#!/usr/bin/env bats

IPVSCTL="$(dirname $BATS_TEST_FILENAME)/../release/ipvsctl"

@test "empty ipvs table yields empty yaml" {
	run $IPVSCTL get 
	[ "$status" -eq 0 ]
	[ "$output" == "{}" ]
}

@test "showing a single service (tcp, rr)" {
	ipvsadm -A -t 1.2.3.4:10000 -s rr
	run $IPVSCTL get 
	ipvsadm -D -t 1.2.3.4:10000

	[ "$status" -eq 0 ]
	[[ "${lines[0]}" == "services:" ]]
	[[ "${lines[1]}" =~ ^-\ address:\ tcp://1.2.3.4:10000$ ]]
	[[ "$output" =~ .*sched:\ rr.* ]]
	[[ ! "$output" =~ destinations ]]
}

@test "showing a single service (tcp, wrr)" {
	ipvsadm -A -t 1.2.3.4:10000 -s wrr
	run $IPVSCTL get 
	ipvsadm -D -t 1.2.3.4:10000

	[ "$status" -eq 0 ]
	[[ "${lines[0]}" == "services:" ]]
	[[ "${lines[1]}" =~ ^-\ address:\ tcp://1.2.3.4:10000$ ]]
	[[ "$output" =~ .*sched:\ wrr.* ]]
	[[ ! "$output" =~ destinations ]]
}

@test "showing a single service (udp, rr)" {
	ipvsadm -A -u 1.2.3.4:10000 -s rr
	run $IPVSCTL get 
	ipvsadm -D -u 1.2.3.4:10000

	[ "$status" -eq 0 ]
	[[ "${lines[0]}" == "services:" ]]
	[[ "${lines[1]}" =~ ^-\ address:\ udp://1.2.3.4:10000$ ]]
	[[ "$output" =~ .*sched:\ rr.* ]]
	[[ ! "$output" =~ destinations ]]
}

@test "showing a single service (fwmark, rr)" {
	ipvsadm -A -f 746 -s rr
	run $IPVSCTL get 
	ipvsadm -D -f 746

	[ "$status" -eq 0 ]
	[[ "${lines[0]}" == "services:" ]]
	[[ "${lines[1]}" =~ ^-\ address:\ fwmark:746$ ]]
	[[ "$output" =~ .*sched:\ rr.* ]]
	[[ ! "$output" =~ destinations ]]
}

@test "showing two services" {
	ipvsadm -A -t 1.2.3.4:10000 -s rr
	ipvsadm -A -t 5.6.7.8:10001 -s rr
	run $IPVSCTL get 
	ipvsadm -D -t 1.2.3.4:10000
	ipvsadm -D -t 5.6.7.8:10001

	[ "$status" -eq 0 ]
	[[ "${lines[0]}" == "services:" ]]
	[[ "$output" =~ address:\ tcp://1.2.3.4:10000 ]]
	[[ "$output" =~ address:\ tcp://5.6.7.8:10001 ]]
	[[ ! "$output" =~ address:\ tcp://5.6.7.8:10000 ]]
	[[ ! "$output" =~ address:\ udp://5.6.7.8:10001 ]]
	[[ ! "$output" =~ destinations ]]
}

# destinations

@test "showing a single service with a single destination (tcp,rr,masq,def.weight)" {
	ipvsadm -A -t 1.2.3.4:10000 -s rr
	ipvsadm -a -t 1.2.3.4:10000 -r 10.0.0.1:20000 -m
	run $IPVSCTL get
	ipvsadm -D -t 1.2.3.4:10000

	[ "$status" -eq 0 ]
	[[ "${lines[0]}" == "services:" ]]
	[[ "${lines[1]}" =~ ^-\ address:\ tcp://1.2.3.4:10000$ ]]
	[[ "$output" =~ .*sched:\ rr.* ]]
	[[ "$output" =~ destinations ]]
	[[ "$output" =~ \ \ -\ address:\ 10.0.0.1:20000 ]]
	[[ ! "$output" =~ \ \ -\ address:\ 10.0.0.2:20000 ]]
	[[ "$output" =~ .*weight:\ 1.* ]]
	# todo: masq
}

@test "showing a single service with two destinations (tcp,rr,masq,def.weight)" {
	ipvsadm -A -t 1.2.3.4:10000 -s rr
	ipvsadm -a -t 1.2.3.4:10000 -r 10.0.0.1:20000 -m
	ipvsadm -a -t 1.2.3.4:10000 -r 10.0.0.2:20000 -m
	run $IPVSCTL get
	ipvsadm -D -t 1.2.3.4:10000

	[ "$status" -eq 0 ]
	[[ "${lines[0]}" == "services:" ]]
	[[ "${lines[1]}" =~ ^-\ address:\ tcp://1.2.3.4:10000$ ]]
	[[ "$output" =~ .*sched:\ rr.* ]]
	[[ "$output" =~ destinations ]]
	[[ "$output" =~ \ \ -\ address:\ 10.0.0.1:20000 ]]
	[[ "$output" =~ \ \ -\ address:\ 10.0.0.2:20000 ]]
	[[ "$output" =~ .*weight:\ 1.* ]]
	# todo: masq
}