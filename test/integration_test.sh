#!/bin/bash

RC=0
set -eo pipefail
cd "$(dirname "$0")"

echo "Starting container"
echo "::group::Docker output"
docker run --rm -d --name tacacs -p 49:49 -v "$(pwd)/tac_plus.cfg:/etc/tac_plus/tac_plus.cfg:ro" lfkeitel/tacacs_plus:alpine
echo "::endgroup::"

echo "Starting tacacs-exporter"
../tacacs-exporter -config "$(pwd)/test_config.yml" &
PID=$!
sleep 1

# test permit
echo "Test 1: Auth Pass"
RESP=$(curl --fail --silent --show-error "http://localhost:9949/metrics?target=127.0.0.1:49&module=success")

echo "::group::Exporter Output"
echo "$RESP"
echo "::endgroup::"

set +e
echo "$RESP" | grep -q "^tacacs_success 1"
RET=$?
if [ $RET -ne 0 ]; then
    echo "::error::tacacs_success 1."
    RC=$RET
fi
echo "$RESP" | grep -q "^tacacs_status_code 1"
RET=$?
if [ $RET -ne 0 ]; then
    echo "::error::tacacs_status_code wasn't 1."
    RC=$RET
fi

# test reject
echo "Test 2: AuthFail(2)"
RESP=$(curl --fail --silent --show-error "http://localhost:9949/metrics?target=127.0.0.1:49&module=reject")

echo "::group::Exporter Output"
echo "$RESP"
echo "::endgroup::"

echo "$RESP" | grep -q "^tacacs_success 1"
RET=$?
if [ $RET -ne 0 ]; then
    echo "::error::tacacs_success wasn't 1."
    RC=$RET
fi
set +e
echo "$RESP" | grep -q "^tacacs_status_code 2"
set -e
RET=$?
if [ $RET -ne 0 ]; then
    echo "::error::tacacs_status_code wasn't 2."
    RC=$RET
fi

echo "Killing tacacs-exporter"
kill $PID

echo "Killing tacacs container"
docker rm -f tacacs

exit $RC