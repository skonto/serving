#!/usr/bin/env bash

# Copyright 2020 The Knative Authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

source $(dirname $0)/../../../test/e2e-common.sh


# Script entry point.
initialize --num-nodes=4 --cluster-version=1.28 "$@"

CERTIFICATE_CLASS="cert-manager.certificate.networking.knative.dev"
toggle_feature external-domain-tls Enabled config-network
restart_pod ${SYSTEM_NAMESPACE} "app=controller"

# Certificate conformance tests must be run separately
# because they need cert-manager specific configurations.
kubectl apply -f "${E2E_YAML_DIR}"/test/config/externaldomaintls/certmanager/selfsigned/
add_trap "kubectl delete -f ./test/e2e/certmanager/config/autotls/certmanager/selfsigned/ --ignore-not-found" SIGKILL SIGTERM SIGQUIT
go_test_e2e -timeout=10m ./test/e2e/certmanager/conformance \
  -run TestNonHTTP01Conformance \
  "--certificateClass=${CERTIFICATE_CLASS}" || fail_test
kubectl delete -f "${E2E_YAML_DIR}"/test/config/externaldomaintls/certmanager/selfsigned/

kubectl apply -f "${E2E_YAML_DIR}"/test/config/externaldomaintls/certmanager/http01/
add_trap "kubectl delete -f ./test/e2e/certmanager/config/autotls/certmanager/http01/ --ignore-not-found" SIGKILL SIGTERM SIGQUIT
go_test_e2e -timeout=10m ./test/e2e/certmanager/conformance \
  -run TestHTTP01Conformance \
  "--certificateClass=${CERTIFICATE_CLASS}" || fail_test
kubectl delete -f "${E2E_YAML_DIR}"/test/config/externaldomaintls/certmanager/http01/

toggle_feature external-domain-tls Disabled config-network
restart_pod ${SYSTEM_NAMESPACE} "app=controller"
success
