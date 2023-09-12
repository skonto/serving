package performance

import (
	"os"
	"strconv"

	"github.com/cloud-bulldozer/go-commons/indexers"
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

	if v, b := os.LookupEnv(UseESEnv); b {
		if b, err = strconv.ParseBool(v); err != nil {
			if b {
				reporter, err = NewESReporter(tags, indexers.ElasticIndexer, index)
				if err != nil {
					return nil, err
				}
			}
		}
	} else if v, b = os.LookupEnv(UseOpenSearcnEnv); b {
		if b, err = strconv.ParseBool(v); err != nil {
			if b {
				reporter, err = NewESReporter(tags, indexers.OpenSearchIndexer, index)
				if err != nil {
					return nil, err
				}
			}
		}
	} else {
		reporter, err = NewInfluxReporter(tags)
		if err != nil {
			return nil, err
		}
	}

	rep := interface{}(reporter).(DataPointReporter)
	return rep, nil
}
