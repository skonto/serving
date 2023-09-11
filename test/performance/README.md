# Performance tests

Knative performance tests are tests geared towards producing useful performance
metrics of the knative system. As such they can choose to take a closed-box
point-of-view of the system and use it just like an end-user might see it. They
can also go more open-box to narrow down the components under test.

## Load Generator

Knative uses [vegeta](https://github.com/tsenart/vegeta/) to generate HTTP load.
It can be configured to generate load at a predefined rate. Officially it
supports constant rate and sine rate, but if you want to generate load at a
different rate, you can write your own pacer by implementing
[Pacer](https://github.com/tsenart/vegeta/blob/e04d9c0df8177e8633bff4afe7b39c2f3a9e7dea/lib/pacer.go#L10)
interface. Custom pacer implementations used in Knative tests are under
[pacers](https://github.com/knative/pkg/tree/main/test/vegeta/pacers).


## Testing architecture

The performance tests are based on Kubernetes Jobs running Golang code based on different [benchmarks](#benchmarks).
The script [performance-tests.sh](./performance-tests.sh) first creates a cluster in GKE, installs Serving with specific settings
for the performance tests. Then it installs the required Knative Services and runs the testing jobs.

The results are written to:
* Stdout
* A logfile in a folder defined in env: `$ARTIFACTS`
* To an InfluxDB hosted in the knative-community GKE project: https://github.com/knative/infra/tree/main/infra/k8s/shared


## Grafana

To better visualize the test results, Grafana is used to show the results in a dashboard.
The dashboard is defined in [grafana-dashboard.json](./visualization/grafana-dashboard.json) and
hosted on [grafana.knative.dev](https://grafana.knative.dev/d/igHJ5-fdk/knative-serving-performance-tests?orgId=1)


## Benchmarks

Knative Serving has different benchmarking scenarios:

* [dataplane-probe](./benchmarks/dataplane-probe): Measures the overhead Knative has compared to a `Deployment`
* [load-test](./benchmarks/load-test): Measures request metrics for Knative Services under load in different scenarios (Activator always in path, Activator only in path at zero, Activator moving out of path on high-load)
* [real-traffic-test](./benchmarks/real-traffic-test): Simulates realistic traffic with random request latency, service startup latency and payload sizes.
* [reconciliation-delay](./benchmarks/reconciliation-delay): Measures the time it takes to reconcile a `KnativeService` and its child CRs.
* [rollout-probe](./benchmarks/rollout-probe): Measures request metrics for a rolling update of a scaled `KnativeService`.
* [scale-from-zero](./benchmarks/scale-from-zero): Measures the latency of scaling 1, 5, 25 and 100 Knative Services from zero in parallel.


## Running the benchmarks

### Local InfluxDB setup

You first need a local running instance of InfluxDB.

Note: if you don't have or don't want to use helm, you can also [install InfluxDB with YAMLs](https://docs.influxdata.com/influxdb/v2.7/install/?t=Kubernetes).

```bash
# Create an InfluxDB with helm
helm repo add influxdata https://helm.influxdata.com/
kubectl create ns influx
helm upgrade --install -n influx local-influx --set persistence.enabled=true,persistence.size=50Gi influxdata/influxdb2
echo "Admin password"
echo $(kubectl get secret local-influx-influxdb2-auth -o "jsonpath={.data['admin-password']}" --namespace influx | base64 --decode)
echo "Admin token"
echo $(kubectl get secret local-influx-influxdb2-auth -o "jsonpath={.data['admin-token']}" --namespace influx | base64 --decode)

# Forward the InfluxDB service to your laptop if you want to access the UI:
kubectl port-forward -n influx svc/local-influx-influxdb2 8080:80

# Set up the expected influxdb config
export INFLUX_URL=http://localhost:8080
export INFLUX_TOKEN=$(kubectl get secret local-influx-influxdb2-auth -o "jsonpath={.data['admin-token']}" --namespace influx | base64 --decode)

# Run the script to initialize the Organization + Buckets in InfluxDB
./visualization/setup-influx-db.sh
```

### Elastic Search / Opensearch

```bash
helm repo add opensearch https://opensearch-project.github.io/helm-charts/
helm repo update
helm search repo opensearch
helm install my-deployment opensearch/opensearch
helm install dashboards opensearch/opensearch-dashboards

export ES_URL=https://localhost:9200
oc port-forward svc/opensearch-cluster-master  9200:9200
export ES_USERNAME=admin
export ES_PASSWORD=admin

# Creates an index template
./test/performance/visualization/setup-es.sh

# Sample log entry to create:

curl -u elastic:Y*A_Ce=+0wbsV8C-b+u* -k -X POST "https://localhost:9200/knative-serving-data-plane/_doc" -H 'Content-Type: application/json' -d'
{
  "@timestamp": "2023-09-12T11:23:23+03:00",
  "_measurement": "Knative Serving dataplane probe",
  "activator-pod-count": "2",
  "tags": [{"JOB_NAME": "custom", "t21":"t2"}]
}
'

SYSTEM_NAMESPACE=knative-serving
JOB_NAME="local"
BUILD_ID="local"
USE_OPEN_SEARCH=true
export ES_URL=https://admin:admin@opensearch-cluster-master.default.svc.cluster.local:9200


kubectl create secret generic performance-test-config -n "default" \
  --from-literal=esurl="${ES_URL}" \
  --from-literal=esusername="${ES_USERNAME}" \
  --from-literal=espassword="${ES_PASSWORD}" \
  --from-literal=prowtag="${JOB_NAME}" \
  --from-literal=buildid="${BUILD_ID}"

ko apply -f ./test/performance/benchmarks/dataplane-probe/dataplane-probe-setup.yaml
sed "s|@SYSTEM_NAMESPACE@|$SYSTEM_NAMESPACE|g" ./test/performance/benchmarks/dataplane-probe/dataplane-probe-deployment.yaml | sed "s|@KO_DOCKER_REPO@|$KO_DOCKER_REPO|g" | sed "s|@USE_OPEN_SEARCH@|\"$USE_OPEN_SEARCH\"|g" | sed "s|@USE_ES@|'false'|g" | ko apply --sbom=none -Bf -
```

### Local grafana dashboards

Use an existing grafana instance or create one on your cluster, [see docs](https://grafana.com/docs/grafana/latest/setup-grafana/installation/kubernetes/).

To use our InfluxDB as a datasource for Grafana
* Navigate to Grafana UI and log in using the user from the installation
* Create a new datasource for InfluxDB
* Select the flux query language
* Server-URL: http://local-influx-influxdb2.influx:80  (Note: this could be different if your grafana instance is hosted outside the cluster)
* Organization: Knativetest
* Bucket: knative-serving
* Token: <your influx-db token>


### Local development

You can run all the benchmarks directly by calling the `main()` method in `main.go` in the respective [benchmarks](./benchmarks) folders.

#### Environment

The tests expect to be configured with certain environment variables:

* KO_DOCKER_REPO = What you have set for `ko`
* SYSTEM_NAMESPACE = Where knative-serving is installed, typically `knative-serving`
* INFLUX_URL = http://local-influx-influxdb2.influx:80
* INFLUX_TOKEN = as outputted from the command above
* BUILD_ID=local
* JOB_NAME=local

### Running them on cluster

Check out what the [script](./performance-tests.sh) does. Basically just run:

```bash
sed "s|@SYSTEM_NAMESPACE@|$SYSTEM_NAMESPACE|g" <your-benchmark-job.yaml> | sed "s|@KO_DOCKER_REPO@|$KO_DOCKER_REPO|g" | ko apply --sbom=none -Bf -
```
