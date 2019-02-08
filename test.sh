#!/bin/sh

TESTS="conn_active \
    conn_passive \
    conn_noip \
    conn_pwd \
    conn_compression \
    chat_public \
    chat_private \
    share \
    tth \
    search \
    download_plain \
    download_tls"

run_test() {
    echo "[$1]"

    docker container kill dctk-verlihub dctk-test >/dev/null 2>&1

    docker run --rm -d --network=dctk-test --name=dctk-verlihub \
        dctk-verlihub ${1} >/dev/null \
        || exit 1

    if [ $VERBOSE -eq 1 ]; then
        docker run --rm -it --network=dctk-test --name=dctk-test \
            -v ${PWD}:/src dctk-test ${1}
        RETCODE=$?
    else
        docker run --rm -it --network=dctk-test --name=dctk-test \
            -v ${PWD}:/src dctk-test ${1} >/dev/null
        RETCODE=$?
    fi

    [ "$RETCODE" -eq 0 ] && echo "RESULT: SUCCESS" || echo "RESULT: FAILED"

    docker container kill dctk-verlihub >/dev/null 2>&1
}

usage() {
    echo "usage: $0 [-v] [all|$(echo $TESTS | tr ' ' '|')]" 1>&2
    exit 1;
}

main() {
    # setup
    docker network rm dctk-test >/dev/null 2>&1; \
        docker build test/verlihub -t dctk-verlihub \
        && docker build . -f test/Dockerfile -t dctk-test \
        && docker network create dctk-test >/dev/null \
        || exit 1

    # read arguments
    VERBOSE=0
    TEST=""
    while [ $# -gt 0 ]; do
        case $1 in
            -v) VERBOSE=1;;
            -*) usage;;
            *) TEST=$1;;
        esac
        shift
    done

    # process
    if [ "$TEST" = "all" ]; then
        for T in $TESTS; do
            run_test $T
        done
    else
        # ensure test exists
        [ -n "$TEST" ] && echo $TESTS | grep -w $TEST >/dev/null \
            || usage

        run_test $TEST
    fi

    # cleanup
    docker network rm dctk-test >/dev/null 2>&1
}

main $@
