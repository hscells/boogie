package main

import (
	"github.com/hscells/groove/query"
	"github.com/hscells/groove/stats"
	"github.com/hscells/groove/analysis"
	"github.com/hscells/groove/output"
)

// RegisterSources initiates boogie with all the possible options in a pipeline.
func RegisterSources() {
	// Query sources.
	RegisterQuerySource("pubmed", query.NewTransmuteQuerySource(query.PubMedTransmutePipeline))
	RegisterQuerySource("medline", query.NewTransmuteQuerySource(query.MedlineTransmutePipeline))

	// Statistic sources.
	RegisterStatisticSource("elasticsearch", stats.NewElasticsearchStatisticsSource())

	// Measurement sources.
	RegisterMeasurement("term_count", analysis.TermCount{})
	RegisterMeasurement("keyword_count", analysis.KeywordCount{})
	RegisterMeasurement("boolean_query_count", analysis.BooleanQueryCount{})

	// Output formats.
	RegisterOutput("json", output.JsonFormatter)
	RegisterOutput("csv", output.CsvFormatter)
}
