package main

import (
	"github.com/hscells/groove/analysis"
	"github.com/hscells/groove/output"
	"github.com/hscells/groove/query"
	"github.com/hscells/groove/stats"
	"github.com/hscells/groove/preprocess"
	"github.com/hscells/transmute/pipeline"
)

var (
	querySourceMapping     = map[string]query.QueriesSource{}
	statisticSourceMapping = map[string]stats.StatisticsSource{}
	preprocessorMapping    = map[string]preprocess.QueryProcessor{}
	transformationMapping  = map[string]preprocess.Transformation{}
	measurementMapping     = map[string]analysis.Measurement{}
	outputMapping          = map[string]output.Formatter{}
)

// RegisterQuerySource registers a query source.
func RegisterQuerySource(name string, source query.QueriesSource) {
	querySourceMapping[name] = source
}

// RegisterStatisticSource registers a statistic source.
func RegisterStatisticSource(name string, source stats.StatisticsSource) {
	statisticSourceMapping[name] = source
}

// RegisterStatisticSource registers a statistic source.
func RegisterPreprocessor(name string, preprocess preprocess.QueryProcessor) {
	preprocessorMapping[name] = preprocess
}

func RegisterTransformation(name string, transformation preprocess.Transformation) {
	transformationMapping[name] = transformation
}

// RegisterMeasurement registers a measurement.
func RegisterMeasurement(name string, measurement analysis.Measurement) {
	measurementMapping[name] = measurement
}

// RegisterOutput registers an output formatter.
func RegisterOutput(name string, formatter output.Formatter) {
	outputMapping[name] = formatter
}

func NewTransmuteQuerySource(p pipeline.TransmutePipeline, options map[string]interface{}) query.QueriesSource {
	if _, ok := options["mapping"]; ok {
		if mapping, ok := options["mapping"].(map[string][]string); ok {
			p.Options.FieldMapping = mapping
		}
	}

	return query.NewTransmuteQuerySource(p)
}

// NewElasticsearchStatisticsSource attempts to create an Elasticsearch statistics source from a configuration mapping.
// It also tries to set some defaults for fields in case some are not specified, but they will not be sensible.
func NewElasticsearchStatisticsSource(config map[string]interface{}) *stats.ElasticsearchStatisticsSource {
	esHosts := []string{}
	documentType := "doc"
	index := "index"
	field := "text"

	if hosts, ok := config["hosts"]; ok {
		for _, host := range hosts.([]interface{}) {
			esHosts = append(esHosts, host.(string))
		}
	} else {
		esHosts = []string{"http://localhost:9200"}
	}

	var searchOptions stats.SearchOptions
	if search, ok := config["search"].(map[string]interface{}); ok {
		if size, ok := search["size"].(int); ok {
			searchOptions.Size = size
		} else {
			searchOptions.Size = 1000
		}

		if runName, ok := search["run_name"].(string); ok {
			searchOptions.RunName = runName
		} else {
			searchOptions.RunName = "run"
		}
	}

	params := map[string]float64{"k": 10, "lambda": 0.5}
	if p, ok := config["params"].(map[string]float64); ok {
		params = p
	}

	if d, ok := config["document_type"]; ok {
		documentType = d.(string)
	}

	if i, ok := config["index"]; ok {
		index = i.(string)
	}

	if f, ok := config["field"]; ok {
		field = f.(string)
	}

	return stats.NewElasticsearchStatisticsSource(
		stats.ElasticsearchDocumentType(documentType),
		stats.ElasticsearchIndex(index),
		stats.ElasticsearchField(field),
		stats.ElasticsearchHosts(esHosts...),
		stats.ElasticsearchParameters(params),
		stats.ElasticsearchSearchOptions(searchOptions))
}
