
# Example run

```bash
export KNATIVE_SERVING_PERF_TEST_DATAPLANE_PROBE_DEPLOYMENT=docker.io/skonto/dataplane-probe-13daa01eab9bbd0b55b029ef0217990f@sha256:5c4b872b0178fee9629dbf46b32bf2a647373943fc7c9daaeaa47870c293e45b
export KNATIVE_SERVING_PERF_TEST_REAL_TRAFFIC=docker.io/skonto/real-traffic-test-0d6cfd702f7100116b002498a1c9d449@sha256:a00a1c9e1a956740b2e1a62bed0792b6975185dd9484da6ebc2263e8b8783594
export KNATIVE_SERVING_PERF_TEST_ROLLOUT_PROBE=docker.io/skonto/rollout-probe-16b878ae522fca2c6d0a486b4be446cd@sha256:be9182f0e5a531f3aa9b7d989675ef387caf1dd65cb05e1a66d13312482edb7d
export KNATIVE_SERVING_PERF_TEST_LOAD_TEST=docker.io/skonto/load-test-16ad8813e1e519c16903611ab3798c1c@sha256:dabbbccb1b30135117b02efa8eb5affd9201dbefb9fa628eb56157bd87759efd
export KNATIVE_SERVING_PERF_TEST_RECONCILIATION_DELAY=docker.io/skonto/reconciliation-delay-6074d88fac79c5d2be9fb1c4ae840488@sha256:b944948d8ebdca4d7109b9dd39a05033324f76279c82f490ec952d0f88924be3
export KNATIVE_SERVING_PERF_TEST_SCALE_FROM_ZERO=docker.io/skonto/scale-from-zero-9924dc8c7b18ccca4da8563a28b55a50@sha256:a823d8b8e0e4669d726f532796e8d907a658a6d393132ddeef442b43daa11ec4
export KNATIVE_SERVING_TEST_AUTOSCALE=docker.io/skonto/autoscale-c163c422b72a456bad9aedab6b2d1f13@sha256:02fc725cef343d41d2278322eef4dd697a6a865290f5afd02ff1a39213f4bbcb


export KNATIVE_SERVING_TEST_RUNTIME=docker.io/skonto/runtime-5fa7cf4c043dfad63fa28de0cfa3b264@sha256:ff5aece839ddec959ed4f2e32c61731ac8ea2550f29a63d73bce100a8a4b004e
export KNATIVE_SERVING_TEST_HELLOWORLD=docker.io/skonto/helloworld-edca531b677458dd5cb687926757a480@sha256:0c9589cde631d33be7548bf54b1e4dbd8e15e486bcd640e0a6c986c5bc1038a6
export KNATIVE_SERVING_TEST_SLOWSTART=docker.io/skonto/slowstart-754e95e646a3d72ab225ebdf3a77a410@sha256:c89ce2dc03377593cd63827b22d4a2f0406fd78870b8e2fff773936940a7efb1

export ES_HOST_PORT=opensearch-cluster-master.default.svc.cluster.local:9200
export ES_USERNAME=admin
export ES_PASSWORD=admin
export SYSTEM_NAMESPACE=knative-serving
export USE_OPEN_SEARCH=true

helm repo add opensearch https://opensearch-project.github.io/helm-charts/
helm repo update
helm search repo opensearch
helm install my-deployment opensearch/opensearch
helm install dashboards opensearch/opensearch-dashboards

oc get po 
NAME                                                READY   STATUS    RESTARTS   AGE
dashboards-opensearch-dashboards-5977566fb9-p4v9m   1/1     Running   0          105m
opensearch-cluster-master-0                         1/1     Running   0          105m
opensearch-cluster-master-1                         1/1     Running   0          105m
opensearch-cluster-master-2                         1/1     Running   0          105m


# On a separate terminal
oc port-forward svc/opensearch-cluster-master 9200:9200

export ES_URL=https://localhost:9200

# Creates an index template for the data
./test/performance/visualization/setup-es.sh

unset ES_URL

./run-all-performance-tests.sh
```