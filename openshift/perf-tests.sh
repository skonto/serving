#!/usr/bin/env bash

# shellcheck disable=SC1090
source "$(dirname "$0")/e2e-common.sh"

set -x
env

failed=0

git apply "$(dirname "$0")/performance/patches/*"

go get github.com/elastic/go-elasticsearch/v7@v7.17.10
go get github.com/opensearch-project/opensearch-go@v1.1.0

"$(dirname "${BASH_SOURCE[0]}")/../hack/update-deps.sh"

git add .
git commit -m "openshift perf update"

git apply "$(dirname "$0")/patches/001-object.patch"
git apply "$(dirname "$0")/patches/002-mutemetrics.patch"
git apply "$(dirname "$0")/patches/003-routeretry.patch"

git add .
git commit -m "apply reverted patches"

(( !failed )) && install_knative || failed=1
"$(dirname "$0")/performance/scripts/run-all-performance-tests.sh"
(( failed )) && gather_knative_state
(( failed )) && exit $failed

success