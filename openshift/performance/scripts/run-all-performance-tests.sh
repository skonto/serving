#!/usr/bin/env bash

declare ns

source "$(dirname "$0")/setup.sh"

set -o errexit
set -o nounset
set -o pipefail


header "Scaling cluster"

for name in $(oc get machineset -n openshift-machine-api -o name); do oc scale $name -n openshift-machine-api --replicas=4; done
oc wait --for=jsonpath={.status.availableReplicas}=4 machineset --all -n openshift-machine-api --timeout=-1s

oc patch knativeserving knative-serving \
    -n "${SYSTEM_NAMESPACE}" \
    --type merge --patch '{"metadata": {"annotations": {"serverless.openshift.io/default-enable-http2": "true" }}}'

###############################################################################################
header "Real traffic test"
toggle_feature kubernetes.podspec-init-containers Enabled
sed -i "s,image: .*,image: ${KNATIVE_SERVING_PERF_TEST_REAL_TRAFFIC}," "${SERVING}/test/performance/benchmarks/real-traffic-test/real-traffic-test.yaml"
run_job real-traffic-test "${SERVING}/test/performance/benchmarks/real-traffic-test/real-traffic-test.yaml" ${KNATIVE_SERVING_PERF_TEST_REAL_TRAFFIC%/*}
sleep 100 # wait a bit for the cleanup to be done
toggle_feature kubernetes.podspec-init-containers Disabled
###############################################################################################
header "Dataplane probe: Setup"

pushd "$SERVING"
sed -i "s,image: .*,image: ${KNATIVE_SERVING_TEST_AUTOSCALE}," "${SERVING}/test/performance/benchmarks/dataplane-probe/dataplane-probe-setup.yaml"
oc apply -f "${SERVING}/test/performance/benchmarks/dataplane-probe/dataplane-probe-setup.yaml"
popd
kubectl wait --timeout=60s --for=condition=ready ksvc -n "$ns" --all
kubectl wait --timeout=60s --for=condition=available deploy -n "$ns" deployment

##############################################################################################
header "Dataplane probe: deployment"

sed -i "s,image: .*,image: ${KNATIVE_SERVING_PERF_TEST_DATAPLANE_PROBE_DEPLOYMENT}," "${SERVING}/test/performance/benchmarks/dataplane-probe/dataplane-probe-deployment.yaml"
run_job dataplane-probe-deployment "${SERVING}/test/performance/benchmarks/dataplane-probe/dataplane-probe-deployment.yaml" ${KNATIVE_SERVING_PERF_TEST_DATAPLANE_PROBE_DEPLOYMENT%/*}

# additional clean up
kubectl delete deploy deployment -n "$ns" --ignore-not-found=true
kubectl delete svc deployment -n "$ns" --ignore-not-found=true
kubectl wait --for=delete deploy/deployment --timeout=60s -n "$ns"
kubectl wait --for=delete svc/deployment --timeout=60s -n "$ns"

##############################################################################################
header "Dataplane probe: activator"

sed -i "s,image: .*,image: ${KNATIVE_SERVING_PERF_TEST_DATAPLANE_PROBE_DEPLOYMENT}," "${SERVING}/test/performance/benchmarks/dataplane-probe/dataplane-probe-activator.yaml"
run_job dataplane-probe-activator "${SERVING}/test/performance/benchmarks/dataplane-probe/dataplane-probe-activator.yaml" ${KNATIVE_SERVING_PERF_TEST_DATAPLANE_PROBE_DEPLOYMENT%/*}

# additional clean up
kubectl delete ksvc activator -n "$ns" --ignore-not-found=true
kubectl wait --for=delete ksvc/activator --timeout=60s -n "$ns"

###############################################################################################
header "Dataplane probe: queue proxy"

sed -i "s,image: .*,image: ${KNATIVE_SERVING_PERF_TEST_DATAPLANE_PROBE_DEPLOYMENT}," "${SERVING}/test/performance/benchmarks/dataplane-probe/dataplane-probe-queue.yaml"
run_job dataplane-probe-queue "${SERVING}/test/performance/benchmarks/dataplane-probe/dataplane-probe-queue.yaml" ${KNATIVE_SERVING_PERF_TEST_DATAPLANE_PROBE_DEPLOYMENT%/*}

# additional clean up
kubectl delete ksvc queue-proxy -n "$ns" --ignore-not-found=true
kubectl wait --for=delete ksvc/queue-proxy --timeout=60s -n "$ns"

###############################################################################################
header "Reconciliation delay test"

sed -i "s,image: .*,image: ${KNATIVE_SERVING_PERF_TEST_RECONCILIATION_DELAY}," "${SERVING}/test/performance/benchmarks/reconciliation-delay/reconciliation-delay.yaml"
run_job reconciliation-delay "${SERVING}/test/performance/benchmarks/reconciliation-delay/reconciliation-delay.yaml" ${KNATIVE_SERVING_PERF_TEST_RECONCILIATION_DELAY%/*}
###############################################################################################
header "Scale from Zero test"

sed -i "s,image: .*,image: ${KNATIVE_SERVING_PERF_TEST_SCALE_FROM_ZERO}," "${SERVING}/test/performance/benchmarks/scale-from-zero/scale-from-zero-1.yaml"
run_job scale-from-zero-1 "${SERVING}/test/performance/benchmarks/scale-from-zero/scale-from-zero-1.yaml"  ${KNATIVE_SERVING_PERF_TEST_SCALE_FROM_ZERO%/*}
kubectl delete ksvc -n "$ns" --all --wait --now
sleep 5 # wait a bit for the cleanup to be done

sed -i "s,image: .*,image: ${KNATIVE_SERVING_PERF_TEST_SCALE_FROM_ZERO}," "${SERVING}/test/performance/benchmarks/scale-from-zero/scale-from-zero-5.yaml"
run_job scale-from-zero-5 "${SERVING}/test/performance/benchmarks/scale-from-zero/scale-from-zero-5.yaml" ${KNATIVE_SERVING_PERF_TEST_SCALE_FROM_ZERO%/*}
kubectl delete ksvc -n "$ns" --all --wait --now
sleep 25 # wait a bit for the cleanup to be done

sed -i "s,image: .*,image: ${KNATIVE_SERVING_PERF_TEST_SCALE_FROM_ZERO}," "${SERVING}/test/performance/benchmarks/scale-from-zero/scale-from-zero-25.yaml"
run_job scale-from-zero-25 "${SERVING}/test/performance/benchmarks/scale-from-zero/scale-from-zero-25.yaml" ${KNATIVE_SERVING_PERF_TEST_SCALE_FROM_ZERO%/*}
kubectl delete ksvc -n "$ns" --all --wait --now
sleep 50 # wait a bit for the cleanup to be done

sed -i "s,image: .*,image: ${KNATIVE_SERVING_PERF_TEST_SCALE_FROM_ZERO}," "${SERVING}/test/performance/benchmarks/scale-from-zero/scale-from-zero-100.yaml"
run_job scale-from-zero-100 "${SERVING}/test/performance/benchmarks/scale-from-zero/scale-from-zero-100.yaml" ${KNATIVE_SERVING_PERF_TEST_SCALE_FROM_ZERO%/*}
kubectl delete ksvc -n "$ns" --all --wait --now
sleep 100 # wait a bit for the cleanup to be done

###############################################################################################
header "Load test: Setup"

pushd "$SERVING"
sed -i "s,image: .*,image: ${KNATIVE_SERVING_TEST_AUTOSCALE}," "${SERVING}/test/performance/benchmarks/load-test/load-test-setup.yaml"
oc apply -f "${SERVING}/test/performance/benchmarks/load-test/load-test-setup.yaml"
popd
kubectl wait --timeout=60s --for=condition=ready ksvc -n "$ns" --all

################################################################################################
header "Load test: zero"

sed -i "s,image: .*,image: ${KNATIVE_SERVING_PERF_TEST_LOAD_TEST}," "${SERVING}/test/performance/benchmarks/load-test/load-test-0-direct.yaml"
run_job load-test-zero "${SERVING}/test/performance/benchmarks/load-test/load-test-0-direct.yaml" ${KNATIVE_SERVING_PERF_TEST_LOAD_TEST%/*}

# additional clean up
kubectl delete ksvc load-test-zero -n "$ns"  --ignore-not-found=true
kubectl wait --for=delete ksvc/load-test-zero --timeout=60s -n "$ns"

##################################################################################################
header "Load test: always direct"

sed -i "s,image: .*,image: ${KNATIVE_SERVING_PERF_TEST_LOAD_TEST}," "${SERVING}/test/performance/benchmarks/load-test/load-test-always-direct.yaml"
run_job load-test-always "${SERVING}/test/performance/benchmarks/load-test/load-test-always-direct.yaml" ${KNATIVE_SERVING_PERF_TEST_LOAD_TEST%/*}

# additional clean up
kubectl delete ksvc load-test-always -n "$ns"  --ignore-not-found=true
kubectl wait --for=delete ksvc/load-test-always --timeout=60s -n "$ns"

#################################################################################################
header "Load test: 200 direct"

sed -i "s,image: .*,image: ${KNATIVE_SERVING_PERF_TEST_LOAD_TEST}," "${SERVING}/test/performance/benchmarks/load-test/load-test-200-direct.yaml"
run_job load-test-200 "${SERVING}/test/performance/benchmarks/load-test/load-test-200-direct.yaml" ${KNATIVE_SERVING_PERF_TEST_LOAD_TEST%/*}

# additional clean up
kubectl delete ksvc load-test-200 -n "$ns"  --ignore-not-found=true
kubectl wait --for=delete ksvc/load-test-200 --timeout=60s -n "$ns"

###############################################################################################
header "Rollout probe: activator direct"

pushd "$SERVING"

sed -i "s,image: .*,image: ${KNATIVE_SERVING_TEST_AUTOSCALE}," "${SERVING}/test/performance/benchmarks/rollout-probe/rollout-probe-setup-activator-direct.yaml"
oc apply -f "${SERVING}/test/performance/benchmarks/rollout-probe/rollout-probe-setup-activator-direct.yaml"
popd
kubectl wait --timeout=800s --for=condition=ready ksvc -n "$ns" --all

sed -i "s,image: .*,image: ${KNATIVE_SERVING_PERF_TEST_ROLLOUT_PROBE}," "${SERVING}/test/performance/benchmarks/rollout-probe/rollout-probe-activator-direct.yaml"
run_job rollout-probe-activator-direct "${SERVING}/test/performance/benchmarks/rollout-probe/rollout-probe-activator-direct.yaml" ${KNATIVE_SERVING_PERF_TEST_ROLLOUT_PROBE%/*}

# additional clean up
kubectl delete ksvc activator-with-cc -n "$ns" --ignore-not-found=true
kubectl wait --for=delete ksvc/activator-with-cc --timeout=60s -n "$ns"

#################################################################################################
header "Rollout probe: activator direct lin"

pushd "$SERVING"
sed -i "s,image: .*,image: ${KNATIVE_SERVING_TEST_AUTOSCALE}," "${SERVING}/test/performance/benchmarks/rollout-probe/rollout-probe-setup-activator-direct-lin.yaml"
oc apply -f "${SERVING}/test/performance/benchmarks/rollout-probe/rollout-probe-setup-activator-direct-lin.yaml"
popd
kubectl wait --timeout=800s --for=condition=ready ksvc -n "$ns" --all

sed -i "s,image: .*,image: ${KNATIVE_SERVING_PERF_TEST_ROLLOUT_PROBE}," "${SERVING}/test/performance/benchmarks/rollout-probe/rollout-probe-activator-direct-lin.yaml"
run_job rollout-probe-activator-direct-lin "${SERVING}/test/performance/benchmarks/rollout-probe/rollout-probe-activator-direct-lin.yaml"  ${KNATIVE_SERVING_PERF_TEST_ROLLOUT_PROBE%/*}

# additional clean up
kubectl delete ksvc activator-with-cc-lin -n "$ns" --ignore-not-found=true
kubectl wait --for=delete ksvc/activator-with-cc-lin --timeout=60s -n "$ns"

##################################################################################################
header "Rollout probe: queue-proxy direct"

pushd "$SERVING"
sed -i "s,image: .*,image: ${KNATIVE_SERVING_TEST_AUTOSCALE}," "${SERVING}/test/performance/benchmarks/rollout-probe/rollout-probe-setup-queue-proxy-direct.yaml"
oc apply -f "${SERVING}/test/performance/benchmarks/rollout-probe/rollout-probe-setup-queue-proxy-direct.yaml"
popd
kubectl wait --timeout=800s --for=condition=ready ksvc -n "$ns" --all

sed -i "s,image: .*,image: ${KNATIVE_SERVING_PERF_TEST_ROLLOUT_PROBE}," "${SERVING}/test/performance/benchmarks/rollout-probe/rollout-probe-queue-proxy-direct.yaml"
run_job rollout-probe-queue-direct "${SERVING}/test/performance/benchmarks/rollout-probe/rollout-probe-queue-proxy-direct.yaml"  ${KNATIVE_SERVING_PERF_TEST_ROLLOUT_PROBE%/*}

# additional clean up
kubectl delete ksvc queue-proxy-with-cc -n "$ns" --ignore-not-found=true
kubectl wait --for=delete ksvc/queue-proxy-with-cc --timeout=60s -n "$ns"

success
