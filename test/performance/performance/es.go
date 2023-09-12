package performance

import (
	indexers2 "knative.dev/serving/test/performance/performance/indexers"
	"log"
	"os"
	"strings"
	"time"

	vegeta "github.com/tsenart/vegeta/v12/lib"
)

const (
	ESServerURLSEnv  = "ES_SERVER_URLS"
	UseOpenSearcnEnv = "USE_OPEN_SEARCH"
	UseESEnv         = "USE_ES"
)

// ESReporter wraps an ES based indexer
type ESReporter struct {
	access *indexers2.Indexer
	tags   map[string]string
}

func sanitizeIndex(index string) string {
	return strings.Replace(index, " ", "_", -1)
}

func splitServers(envURLS string) []string {
	var addrs []string
	list := strings.Split(envURLS, ",")
	for _, u := range list {
		addrs = append(addrs, strings.TrimSpace(u))
	}
	return addrs
}

func NewESReporter(tags map[string]string, indexerType indexers2.IndexerType, index string) (*ESReporter, error) {
	var servers []string

	if v, b := os.LookupEnv(ESServerURLSEnv); b {
		servers = splitServers(v)
	}
	indexer, err := indexers2.NewIndexer(indexers2.IndexerConfig{
		Type:               indexerType,
		Index:              sanitizeIndex(index),
		Servers:            servers,
		InsecureSkipVerify: true,
	})
	if err != nil {
		return nil, err
	}
	return &ESReporter{
		access: indexer,
		tags:   tags,
	}, nil
}

func (esr *ESReporter) AddDataPointsForMetrics(m *vegeta.Metrics, benchmarkName string) {
	metrics := []map[string]interface{}{
		{
			"requests":     float64(m.Requests),
			"rate":         m.Rate,
			"throughput":   m.Throughput,
			"duration":     float64(m.Duration),
			"latency-mean": float64(m.Latencies.Mean),
			"latency-min":  float64(m.Latencies.Min),
			"latency-max":  float64(m.Latencies.Max),
			"latency-p95":  float64(m.Latencies.P95),
			"success":      m.Success,
			"errors":       float64(len(m.Errors)),
			"bytes-in":     float64(m.BytesIn.Total),
			"bytes-out":    float64(m.BytesOut.Total),
		},
	}

	for _, m := range metrics {
		esr.AddDataPoint(benchmarkName, m)
	}
}

func (esr *ESReporter) AddDataPoint(measurement string, fields map[string]interface{}) {
	p := fields
	p["_measurement"] = measurement
	p["tags"] = esr.tags
	// Use the same format as in influxdb
	p["@timestamp"] = time.Now().Format(time.RFC3339Nano)
	docs := []interface{}{p}
	msg, err := (*esr.access).Index(docs, indexers2.IndexingOpts{})
	if err != nil {
		log.Printf("Indexing failed: %s", err.Error())
	}
	log.Printf("%s\n", msg)
}

func (esr *ESReporter) FlushAndShutdown() {

}
