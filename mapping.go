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
	querySourceMapping                 = map[string]query.QueriesSource{}
	statisticSourceMapping             = map[string]stats.StatisticsSource{}
	preprocessorMapping                = map[string]preprocess.QueryProcessor{}
	transformationMappingBoolean       = map[string]preprocess.BooleanTransformation{}
	transformationMappingElasticsearch = map[string]preprocess.ElasticsearchTransformation{}
	measurementMapping                 = map[string]analysis.Measurement{}
	outputMapping                      = map[string]output.Formatter{}
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

func RegisterTransformationBoolean(name string, transformation preprocess.BooleanTransformation) {
	transformationMappingBoolean[name] = transformation
}

func RegisterTransformationElasticsearch(name string, transformation preprocess.ElasticsearchTransformation) {
	transformationMappingElasticsearch[name] = transformation
}

// RegisterMeasurement registers a measurement.
func RegisterMeasurement(name string, measurement analysis.Measurement) {
	measurementMapping[name] = measurement
}

// RegisterOutput registers an output formatter.
func RegisterOutput(name string, formatter output.Formatter) {
	outputMapping[name] = formatter
}

func NewKeywordQuerySource(options map[string]interface{}) query.QueriesSource {
	fields := []string{"text"}
	if optionFields, ok := options["fields"].([]interface{}); ok {
		fields = make([]string, len(optionFields))
		for i, f := range optionFields {
			fields[i] = f.(string)
		}
	}

	return query.NewKeywordQuerySource(fields...)
}

func NewTransmuteQuerySource(p pipeline.TransmutePipeline, options map[string]interface{}) query.QueriesSource {
	if _, ok := options["mapping"]; ok {
		if mapping, ok := options["mapping"].(map[string][]string); ok {
			p.Options.FieldMapping = mapping
		}
	}

	return query.NewTransmuteQuerySource(p)
}

// NewTerrierStatisticsSource attempts to create a terrier statistics source.
func NewTerrierStatisticsSource(config map[string]interface{}) *stats.TerrierStatisticsSource {
	var propsFile string
	field := "text"

	if pf, ok := config["properties"]; ok {
		propsFile = pf.(string)
	}

	if f, ok := config["field"]; ok {
		field = f.(string)
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

	return stats.NewTerrierStatisticsSource(stats.TerrierParameters(params), stats.TerrierField(field), stats.TerrierPropertiesPath(propsFile), stats.TerrierSearchOptions(searchOptions))
}

// NewElasticsearchStatisticsSource attempts to create an Elasticsearch statistics source from a configuration mapping.
// It also tries to set some defaults for fields in case some are not specified, but they will not be sensible.
func NewElasticsearchStatisticsSource(config map[string]interface{}) *stats.ElasticsearchStatisticsSource {
	esHosts := []string{}
	documentType := "doc"
	index := "index"
	field := "text"
	analyser := "standard"

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

	if a, ok := config["analyser"]; ok {
		analyser = a.(string)
	}

	return stats.NewElasticsearchStatisticsSource(
		stats.ElasticsearchDocumentType(documentType),
		stats.ElasticsearchIndex(index),
		stats.ElasticsearchField(field),
		stats.ElasticsearchHosts(esHosts...),
		stats.ElasticsearchParameters(params),
		stats.ElasticsearchAnalyser(analyser),
		stats.ElasticsearchSearchOptions(searchOptions))
}
