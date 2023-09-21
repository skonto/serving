#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

# shellcheck disable=SC1091
source "$(dirname "$0")/setup.sh"

# Customization
frequency=$1
export frequency=$frequency
envsubst < "$(dirname "$0")/../scenarios/customizations/reconciliation-delay.yaml" > "$SERVING/test/performance/benchmarks/reconciliation-delay/reconciliation-delay.yaml"


# Run the tests
run_job reconciliation-delay "$SERVING/test/performance/benchmarks/reconciliation-delay/reconciliation-delay.yaml"
