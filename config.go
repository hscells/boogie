package main

import (
	"github.com/hscells/groove/analysis"
	"github.com/hscells/groove/output"
	"github.com/hscells/groove/query"
	"github.com/hscells/groove/stats"
	"github.com/hscells/groove/analysis/preqpp"
	"github.com/hscells/groove/preprocess"
)

// RegisterSources initiates boogie with all the possible options in a pipeline.
func RegisterSources() {
	// Query sources.
	RegisterQuerySource("pubmed", query.NewTransmuteQuerySource(query.PubMedTransmutePipeline))
	RegisterQuerySource("medline", query.NewTransmuteQuerySource(query.MedlineTransmutePipeline))

	// Statistic sources.
	RegisterStatisticSource("elasticsearch", NewElasticsearchStatisticsSource(dsl.Statistic.Options))

	// Preprocessor sources.
	RegisterPreprocessor("alphanum", preprocess.AlphaNum)
	RegisterPreprocessor("lowercase", preprocess.Lowercase)

	// Measurement sources.
	RegisterMeasurement("term_count", analysis.TermCount{})
	RegisterMeasurement("keyword_count", analysis.KeywordCount{})
	RegisterMeasurement("boolean_query_count", analysis.BooleanQueryCount{})

	RegisterMeasurement("avg_idf", preqpp.AvgIDF{})
	RegisterMeasurement("sum_idf", preqpp.SumIDF{})

	// Output formats.
	RegisterOutput("json", output.JsonFormatter)
	RegisterOutput("csv", output.CsvFormatter)
}

// NewElasticsearchStatisticsSource attempts to create an Elasticsearch statistics source from a configuration mapping.
// It also tries to set some defaults for fields in case some are not specified, but they will not be sensible.
func NewElasticsearchStatisticsSource(config map[string]interface{}) stats.ElasticsearchStatisticsSource {
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

	if d, ok := config["document_type"]; ok {
		documentType = d.(string)
	}

	if i, ok := config["index"]; ok {
		index = i.(string)
	}

	if f, ok := config["field"]; ok {
		field = f.(string)
	}

	return *stats.NewElasticsearchStatisticsSource(
		stats.ElasticsearchHosts(esHosts...),
		stats.ElasticsearchDocumentType(documentType),
		stats.ElasticsearchIndex(index),
		stats.ElasticsearchField(field))
}
