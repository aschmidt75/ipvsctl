#!/usr/bin/env bats

IPVSCTL="$(dirname $BATS_TEST_FILENAME)/../dist/ipvsctl"
if [ ! -x "${IPVSCTL}" ]; then   
    IPVSCTL=$(which ipvsctl)
    if [ ! -x "${IPVSCTL}" ]; then   
        echo ERROR unable to find ipvsctl in local dist or in path
        exit 1
    fi
fi

@test "given a model with a network parameter, when i validate it without parameters enabled, it should fail" {
	run $IPVSCTL validate -f fixtures/params-network.yaml

	[ "$status" -eq 35 ]
}

@test "given a model with a network parameter, when i validate it with network parameters enabled, it should pass" {
	run $IPVSCTL --params-network validate -f fixtures/params-network.yaml

	[ "$status" -eq 0 ]
}

@test "given a model with an environment parameter, when i validate it without parameters enabled, it should fail" {
	run $IPVSCTL validate -f fixtures/params-env.yaml

	[ "$status" -eq 35 ]
}

@test "given a model with an environment parameter, when i validate it with env parameters enabled, it should pass" {
	TESTPORT=80 run $IPVSCTL --params-env validate -f fixtures/params-env.yaml

	[ "$status" -eq 0 ]
}

@test "given a parameterized model, when it validate it with a non-existing parameter file, it should fail" {
    run $IPVSCTL --params-file=no-such-params-file validate -f fixtures/params-fromfile.yaml

	[ "$status" -eq 51 ]
}

@test "given a parameterized model, when it validate it with an invalid parameter file, it should fail" {
    run $IPVSCTL --params-file=fixtures/params/params-file.invalid validate -f fixtures/params-fromfile.yaml

	[ "$status" -eq 30 ]
}

@test "given a parameterized model, when it validate it with an valid yaml parameter file, it should pass" {
    run $IPVSCTL --params-file=fixtures/params/params-file.yaml validate -f fixtures/params-fromfile.yaml

	[ "$status" -eq 0 ]
}

@test "given a parameterized model, when it validate it with an valid json parameter file, it should pass" {
    run $IPVSCTL --params-file=fixtures/params/params-file.json validate -f fixtures/params-fromfile.yaml

	[ "$status" -eq 0 ]
}

@test "given a parameterized model, when it validate it with two valid parameter files, it should pass" {
    run $IPVSCTL --params-file=fixtures/params/params-file.yaml --params-file=fixtures/params/params-file.json validate -f fixtures/params-fromfile2.yaml

	[ "$status" -eq 0 ]
}

@test "given a parameterized model, when it validate it with a missing parameter file, it should fail" {
    run $IPVSCTL --params-file=fixtures/params/params-file.json validate -f fixtures/params-fromfile2.yaml

	[ "$status" -ne 0 ]
}

@test "given a parameterized model, when it validate it with an invalid parameter URL(s), it should fail" {
    run $IPVSCTL --params-url=http://127.0.0.1:7777//no-such-params-file.yaml validate -f fixtures/params-fromfile.yaml
	[ "$status" -ne 0 ]
}

@test "given a parameterized model, when it validate it with an valid parameter URL(s), it should pass" {

#DV=$(which docker 2>/dev/null || true)
#if [ "${DV}" != "" ]; then  
#
#    TESTCONTAINER=$(docker run -p 127.0.0.1:9999:80 -v $(pwd)/fixtures/params:/usr/share/nginx/html:ro -d nginx)
#
#    run $IPVSCTL --params-url=http://127.0.0.1:9999//params-file.yaml validate -f fixtures/params-fromfile.yaml#
#	[ "$status" -eq 0 ]
#
#    run $IPVSCTL --params-url=http://127.0.0.1:9999//params-file.json validate -f fixtures/params-fromfile.yaml
#	[ "$status" -eq 0 ]
#
#    run $IPVSCTL --params-url=http://127.0.0.1:9999//params-file.yaml --params-url=http://127.0.0.1:9999//params-file.json validate -f fixtures/params-fromfile2.yaml
#	[ "$status" -eq 0 ]
#
#    run $IPVSCTL --params-url=http://127.0.0.1:9999//params-file.yaml --params-file=fixtures/params/params-file.json validate -f fixtures/params-fromfile2.yaml
#	[ "$status" -eq 0 ]
#
#    docker stop ${TESTCONTAINER} && docker rm ${TESTCONTAINER}
#fi
}

