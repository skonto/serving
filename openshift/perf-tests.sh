#!/usr/bin/env bash

# shellcheck disable=SC1090
source "$(dirname "$0")/e2e-common.sh"

set -x
env

failed=0

(( !failed )) && install_knative || failed=1
"$(dirname "$0")/performance/scripts/run-all-performance-tests.sh"
(( failed )) && gather_knative_state
(( failed )) && exit $failed

success