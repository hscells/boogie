package boogie

import (
	"github.com/hscells/groove/analysis"
	"github.com/hscells/groove/analysis/postqpp"
	"github.com/hscells/groove/analysis/preqpp"
	"github.com/hscells/groove/eval"
	"github.com/hscells/groove/learning"
	"github.com/hscells/groove/output"
	"github.com/hscells/groove/preprocess"
	"github.com/hscells/groove/query"
	"fmt"
	"errors"
)

// RegisterSources initiates boogie with all the possible options in a pipeline.
func RegisterSources(dsl Pipeline) error {
	// Query sources.
	RegisterQuerySource("medline", NewTransmuteQuerySource(query.MedlineTransmutePipeline, dsl.Query.Options))
	RegisterQuerySource("pubmed", NewTransmuteQuerySource(query.PubMedTransmutePipeline, dsl.Query.Options))
	RegisterQuerySource("keyword", NewKeywordQuerySource(dsl.Query.Options))

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
	RegisterEvaluator("f05_measure", eval.F05Measure)
	RegisterEvaluator("f1_measure", eval.F1Measure)
	RegisterEvaluator("f3_measure", eval.F3Measure)
	// TODO add WSS evaluation metric.

	// Output formats.
	RegisterMeasurementFormatter("json", output.JsonMeasurementFormatter)
	RegisterMeasurementFormatter("csv", output.CsvMeasurementFormatter)
	RegisterEvaluationFormatter("json", output.JsonEvaluationFormatter)

	// Query Rewrite transformations.
	RegisterRewriteTransformation("logical_operator", learning.NewLogicalOperatorTransformer())
	RegisterRewriteTransformation("adj_range", learning.NewAdjacencyRangeTransformer())
	RegisterRewriteTransformation("mesh_explosion", learning.NewMeSHExplosionTransformer())
	RegisterRewriteTransformation("field_restrictions", learning.NewFieldRestrictionsTransformer())
	RegisterRewriteTransformation("adj_replacement", learning.NewAdjacencyReplacementTransformer())
	RegisterRewriteTransformation("clause_removal", learning.NewClauseRemovalTransformer())
	// TODO add cui_expansion transformation.

	// Machine learning models.
	switch m := dsl.Learning.Model; m {
	case "ltr_query_chain":
		RegisterModel(m, learning.NewLearningToRankQueryChain(dsl.Learning.Options["model_file"]))
	case "reinforcement_query_chain":
		RegisterModel(m, learning.NewReinforcementQueryChain())
	default:
		return errors.New(fmt.Sprintf("could not load model of type %s", m))
	}

	// Statistic sources.
	switch s := dsl.Statistic.Source; s {
	case "elasticsearch":
		ss, err := NewElasticsearchStatisticsSource(dsl.Statistic.Options)
		if err != nil {
			return err
		}
		RegisterStatisticSource(s, ss)
	case "terrier":
		RegisterStatisticSource(s, NewTerrierStatisticsSource(dsl.Statistic.Options))
	case "entrez":
		ss, err := NewEntrezStatisticsSource(dsl.Statistic.Options)
		if err != nil {
			return err
		}
		RegisterStatisticSource(s, ss)
	default:
		return errors.New(fmt.Sprintf("could not load statistic source %s", s))
	}

	return nil
}
