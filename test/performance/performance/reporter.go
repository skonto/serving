package performance

import (
	"knative.dev/serving/test/performance/performance/indexers"
	"log"
	"os"
	"strconv"

	vegeta "github.com/tsenart/vegeta/v12/lib"
)

type DataPointReporter interface {
	AddDataPoint(measurement string, fields map[string]interface{})
	AddDataPointsForMetrics(m *vegeta.Metrics, benchmarkName string)
	FlushAndShutdown()
}

func NewDataPointReporterFactory(tags map[string]string, index string) (DataPointReporter, error) {
	var reporter DataPointReporter
	var err error
	useDefaultReporter := true

	if v, b := os.LookupEnv(UseESEnv); b {
		if b, err = strconv.ParseBool(v); err == nil {
			if b {
				useDefaultReporter = false
				reporter, err = NewESReporter(tags, indexers.ElasticIndexer, index)
				if err != nil {
					return nil, err
				}
			}
		}
	}

	if v, b := os.LookupEnv(UseOpenSearcnEnv); b {
		if b, err = strconv.ParseBool(v); err == nil {
			if b {
				useDefaultReporter = false
				log.Println("here 3")
				reporter, err = NewESReporter(tags, indexers.OpenSearchIndexer, index)
				if err != nil {
					log.Println("here 3.1")
					return nil, err
				}
			}
		}
	}

	if useDefaultReporter {
		reporter, err = NewInfluxReporter(tags)
		if err != nil {
			return nil, err
		}
	}

	rep := interface{}(reporter).(DataPointReporter)
	return rep, nil
}
