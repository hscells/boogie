package main

import (
	"github.com/hscells/groove/analysis"
	"github.com/hscells/groove/analysis/postqpp"
	"github.com/hscells/groove/analysis/preqpp"
	"github.com/hscells/groove/output"
	"github.com/hscells/groove/preprocess"
	"github.com/hscells/groove/query"
	"log"
	"github.com/hscells/groove/eval"
)

// RegisterSources initiates boogie with all the possible options in a pipeline.
func RegisterSources() {
	// Query sources.
	RegisterQuerySource("medline", NewTransmuteQuerySource(query.MedlineTransmutePipeline, dsl.Query.Options))
	RegisterQuerySource("pubmed", NewTransmuteQuerySource(query.PubMedTransmutePipeline, dsl.Query.Options))
	RegisterQuerySource("keyword", NewKeywordQuerySource(dsl.Query.Options))

	// Statistic sources.
	switch s := dsl.Statistic.Source; s {
	case "elasticsearch":
		RegisterStatisticSource(s, NewElasticsearchStatisticsSource(dsl.Statistic.Options))
	case "terrier":
		RegisterStatisticSource(s, NewTerrierStatisticsSource(dsl.Statistic.Options))
	default:
		log.Fatalf("could not load statistic source %s", s)
	}

	// Preprocessor sources.
	RegisterPreprocessor("alphanum", preprocess.AlphaNum)
	RegisterPreprocessor("lowercase", preprocess.Lowercase)
	RegisterPreprocessor("strip_numbers", preprocess.StripNumbers)

	// Transformations.
	RegisterTransformationBoolean("simplify", preprocess.Simplify)
	RegisterTransformationElasticsearch("analyse", preprocess.Analyse)

	// Measurement sources.
	RegisterMeasurement("term_count", analysis.TermCount)
	RegisterMeasurement("keyword_count", analysis.KeywordCount)
	RegisterMeasurement("boolean_query_count", analysis.BooleanQueryCount)
	RegisterMeasurement("sum_idf", preqpp.SumIDF)
	RegisterMeasurement("avg_idf", preqpp.AvgIDF)
	RegisterMeasurement("max_idf", preqpp.MaxIDF)
	RegisterMeasurement("std_idf", preqpp.StdDevIDF)
	RegisterMeasurement("avg_ictf", preqpp.AvgICTF)
	RegisterMeasurement("query_scope", preqpp.QueryScope)
	RegisterMeasurement("scs", preqpp.SimplifiedClarityScore)
	RegisterMeasurement("sum_cqs", preqpp.SummedCollectionQuerySimilarity)
	RegisterMeasurement("max_cqs", preqpp.MaxCollectionQuerySimilarity)
	RegisterMeasurement("avg_cqs", preqpp.AverageCollectionQuerySimilarity)
	RegisterMeasurement("wig", postqpp.WeightedInformationGain)
	RegisterMeasurement("weg", postqpp.WeightedExpansionGain)
	RegisterMeasurement("ncq", postqpp.NormalisedQueryCommitment)
	RegisterMeasurement("clarity_score", postqpp.ClarityScore)

	// Evaluations measurements.
	RegisterEvaluator("precision", eval.PrecisionEvaluator)
	RegisterEvaluator("recall", eval.RecallEvaluator)
	RegisterEvaluator("num_rel", eval.NumRel)
	RegisterEvaluator("num_ret", eval.NumRet)
	RegisterEvaluator("num_rel_ret", eval.NumRelRet)

	// Output formats.
	RegisterMeasurementFormatter("json", output.JsonMeasurementFormatter)
	RegisterMeasurementFormatter("csv", output.CsvMeasurementFormatter)
	RegisterEvaluationFormatter("json", output.JsonEvaluationFormatter)

}
