package main

import (
	"github.com/hscells/groove/analysis"
	"github.com/hscells/groove/output"
	"github.com/hscells/groove/query"
	"github.com/hscells/groove/analysis/preqpp"
	"github.com/hscells/groove/preprocess"
	"github.com/hscells/groove/stats"
	"log"
	"github.com/hscells/groove/analysis/postqpp"
)

// RegisterSources initiates boogie with all the possible options in a pipeline.
func RegisterSources() {
	// Query sources.
	RegisterQuerySource("medline", NewTransmuteQuerySource(query.MedlineTransmutePipeline, dsl.Query.Options))
	RegisterQuerySource("pubmed", NewTransmuteQuerySource(query.PubMedTransmutePipeline, dsl.Query.Options))

	// Statistic sources.
	switch s := dsl.Statistic.Source; s {
	case "elasticsearch":
		RegisterStatisticSource(s, NewElasticsearchStatisticsSource(dsl.Statistic.Options))
	case "terrier":
		RegisterStatisticSource("terrier", stats.NewTerrierStatisticsSource(stats.TerrierPropertiesPath("/Users/harryscells/terrier-core-4.2/etc/terrier.properties")))
	default:
		log.Fatalf("could not load statistic source %s", s)
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
	RegisterMeasurement("wig", postqpp.WeightedInformationGain{})
	RegisterMeasurement("weg", postqpp.WeightedExpansionGain{})

	// Output formats.
	RegisterOutput("json", output.JsonFormatter)
	RegisterOutput("csv", output.CsvFormatter)
}
