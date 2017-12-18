package main

import (
	"github.com/hscells/groove/analysis"
	"github.com/hscells/groove/output"
	"github.com/hscells/groove/preprocess"
	"github.com/hscells/groove/query"
	"github.com/hscells/groove/stats"
	"github.com/hscells/transmute/pipeline"
	"github.com/hscells/groove/eval"
)

var (
	querySourceMapping                 = map[string]query.QueriesSource{}
	statisticSourceMapping             = map[string]stats.StatisticsSource{}
	preprocessorMapping                = map[string]preprocess.QueryProcessor{}
	transformationMappingBoolean       = map[string]preprocess.BooleanTransformation{}
	transformationMappingElasticsearch = map[string]preprocess.ElasticsearchTransformation{}
	measurementMapping                 = map[string]analysis.Measurement{}
	measurementFormatters              = map[string]output.MeasurementFormatter{}
	evaluationMapping                  = map[string]eval.Evaluator{}
	evaluationFormatters               = map[string]output.EvaluationFormatter{}
)

// RegisterQuerySource registers a query source.
func RegisterQuerySource(name string, source query.QueriesSource) {
	querySourceMapping[name] = source
}

// RegisterStatisticSource registers a statistic source.
func RegisterStatisticSource(name string, source stats.StatisticsSource) {
	statisticSourceMapping[name] = source
}

// RegisterPreprocessor registers a preprocessor.
func RegisterPreprocessor(name string, preprocess preprocess.QueryProcessor) {
	preprocessorMapping[name] = preprocess
}

// RegisterTransformationBoolean registers a Boolean query transformation.
func RegisterTransformationBoolean(name string, transformation preprocess.BooleanTransformation) {
	transformationMappingBoolean[name] = transformation
}

// RegisterTransformationElasticsearch registers an Elasticsearch transformation.
func RegisterTransformationElasticsearch(name string, transformation preprocess.ElasticsearchTransformation) {
	transformationMappingElasticsearch[name] = transformation
}

// RegisterMeasurement registers a measurement.
func RegisterMeasurement(name string, measurement analysis.Measurement) {
	measurementMapping[name] = measurement
}

// RegisterMeasurementFormatter registers an output formatter.
func RegisterMeasurementFormatter(name string, formatter output.MeasurementFormatter) {
	measurementFormatters[name] = formatter
}

// RegisterEvaluator registers a measurement.
func RegisterEvaluator(name string, evaluator eval.Evaluator) {
	evaluationMapping[name] = evaluator
}

// RegisterMeasurementFormatter registers an output formatter.
func RegisterEvaluationFormatter(name string, formatter output.EvaluationFormatter) {
	evaluationFormatters[name] = formatter
}

// NewKeywordQuerySource creates a "keyword query" query source.
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

// NewTransmuteQuerySource creates a new transmute query source for PubMed/Medline queries.
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
	var esHosts []string
	documentType := "doc"
	index := "index"
	field := "text"
	analyser := "standard"
	analyseField := ""
	scroll := false

	if hosts, ok := config["hosts"]; ok {
		for _, host := range hosts.([]interface{}) {
			esHosts = append(esHosts, host.(string))
		}
	} else {
		esHosts = []string{"http://localhost:9200"}
	}

	var searchOptions stats.SearchOptions
	if search, ok := config["search"].(map[string]interface{}); ok {
		if size, ok := search["size"].(float64); ok {
			searchOptions.Size = int(size)
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

	if a, ok := config["analyse_field"]; ok {
		analyseField = a.(string)
	}

	if s, ok := config["scroll"]; ok {
		scroll = s.(bool)
	}

	return stats.NewElasticsearchStatisticsSource(
		stats.ElasticsearchDocumentType(documentType),
		stats.ElasticsearchIndex(index),
		stats.ElasticsearchField(field),
		stats.ElasticsearchHosts(esHosts...),
		stats.ElasticsearchParameters(params),
		stats.ElasticsearchAnalyser(analyser),
		stats.ElasticsearchAnalysedField(analyseField),
		stats.ElasticsearchSearchOptions(searchOptions),
		stats.ElasticsearchScroll(scroll))
}
