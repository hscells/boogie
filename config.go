package main

import (
	"github.com/hscells/groove/analysis"
	"github.com/hscells/groove/output"
	"github.com/hscells/groove/query"
	"github.com/hscells/groove/analysis/preqpp"
	"github.com/hscells/groove/preprocess"
)

// RegisterSources initiates boogie with all the possible options in a pipeline.
func RegisterSources() {
	// Query sources.
	RegisterQuerySource("medline", NewTransmuteQuerySource(query.MedlineTransmutePipeline, dsl.Query.Options))
	RegisterQuerySource("pubmed", NewTransmuteQuerySource(query.PubMedTransmutePipeline, dsl.Query.Options))

	// Statistic sources.
	if len(dsl.Statistic.Source) > 0 {
		RegisterStatisticSource("elasticsearch", NewElasticsearchStatisticsSource(dsl.Statistic.Options))
	}

	// Preprocessor sources.
	RegisterPreprocessor("alphanum", preprocess.AlphaNum)
	RegisterPreprocessor("lowercase", preprocess.Lowercase)

	// Transformations.
	RegisterTransformation("simplify", preprocess.Simplify)

	// Measurement sources.
	RegisterMeasurement("term_count", analysis.TermCount{})
	RegisterMeasurement("keyword_count", analysis.KeywordCount{})
	RegisterMeasurement("boolean_query_count", analysis.BooleanQueryCount{})
	RegisterMeasurement("sum_idf", preqpp.SumIDF{})
	RegisterMeasurement("avg_idf", preqpp.AvgIDF{})
	RegisterMeasurement("max_idf", preqpp.MaxIDF{})
	RegisterMeasurement("std_idf", preqpp.StdDevIDF{})
	RegisterMeasurement("avg_ictf", preqpp.AvgICTF{})
	RegisterMeasurement("query_scope", preqpp.QueryScope{})
	RegisterMeasurement("scs", preqpp.SimplifiedClarityScore{})
	RegisterMeasurement("sum_cqs", preqpp.SummedCollectionQuerySimilarity{})
	RegisterMeasurement("max_cqs", preqpp.MaxCollectionQuerySimilarity{})

	// Output formats.
	RegisterOutput("json", output.JsonFormatter)
	RegisterOutput("csv", output.CsvFormatter)
}
