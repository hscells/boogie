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
	"github.com/hscells/groove/rewrite"
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
	RegisterTransformationBoolean("and_simplify", preprocess.AndSimplify)
	RegisterTransformationBoolean("or_simplify", preprocess.OrSimplify)
	RegisterTransformationBoolean("rct_filter", preprocess.RCTFilter)
	RegisterTransformationElasticsearch("analyse", preprocess.Analyse)
	RegisterTransformationElasticsearch("set_analyse", preprocess.SetAnalyseField)

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
	RegisterEvaluator("distributedness", eval.Distributedness)

	// Output formats.
	RegisterMeasurementFormatter("json", output.JsonMeasurementFormatter)
	RegisterMeasurementFormatter("csv", output.CsvMeasurementFormatter)
	RegisterEvaluationFormatter("json", output.JsonEvaluationFormatter)

	// Rewrite.
	// Query Chain Candidate Selectors.
	RegisterQueryChainCandidateSelector("oracle", NewOracleQueryChainCandidateSelector(dsl.Statistic.Source, dsl.Output.Evaluations.Qrels))

	// Rewrite Transformations.
	RegisterRewriteTransformation("logical_operator", rewrite.LogicalOperatorReplacement)
	RegisterRewriteTransformation("adj_range", rewrite.AdjacencyRange)
	RegisterRewriteTransformation("mesh_explosion", rewrite.MeSHExplosion)
	RegisterRewriteTransformation("field_restrictions", rewrite.FieldRestrictions)
	RegisterRewriteTransformation("adj_replacement", rewrite.AdjacencyReplacement)
}
