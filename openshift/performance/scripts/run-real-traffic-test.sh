#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

# shellcheck disable=SC2034
declare ns

# shellcheck disable=SC1091
source "$(dirname "$0")/setup.sh"

# Running the tests

################################################################################################
header "Real traffic test"

run_job real-traffic-test "$(dirname "$0")/test/performance/benchmarks/real-traffic-test/real-traffic-test.yaml"
