package boogie

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/hscells/groove/analysis"
	"github.com/hscells/groove/analysis/postqpp"
	"github.com/hscells/groove/analysis/preqpp"
	"github.com/hscells/groove/eval"
	"github.com/hscells/groove/learning"
	"github.com/hscells/groove/output"
	"github.com/hscells/groove/preprocess"
	"github.com/hscells/groove/query"
	"github.com/hscells/trecresults"
	"io/ioutil"
	"os"
	"strconv"
)

// RegisterSources initiates boogie with all the possible options in a pipeline.
func RegisterSources(dsl Pipeline) error {
	// Statistic sources.
	// Configuration of other parts of the pipeline can depend on the statistics source
	// so this needs to be set up first.
	switch s := dsl.Statistic.Source; s {
	case "elasticsearch":
		ss, err := NewElasticsearchStatisticsSource(dsl.Statistic.Options)
		if err != nil {
			return err
		}
		RegisterStatisticSource(s, ss)
	case "terrier":
		// TODO rework code to allow linux to use Terrier.
		//RegisterStatisticSource(s, NewTerrierStatisticsSource(dsl.Statistic.Options))
	case "entrez":
		ss, err := NewEntrezStatisticsSource(dsl.Statistic.Options)
		if err != nil {
			return err
		}
		RegisterStatisticSource(s, ss)
	}

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
	RegisterMeasurement("retrieval_size", preqpp.RetrievalSize)
	RegisterMeasurement("boolean_clauses", analysis.BooleanClauses)
	RegisterMeasurement("boolean_keywords", analysis.BooleanKeywords)
	RegisterMeasurement("boolean_fields", analysis.BooleanFields)
	RegisterMeasurement("boolean_truncated", analysis.BooleanTruncated)
	RegisterMeasurement("boolean_atomicnonatomic", analysis.BooleanAtomicNonAtomic)
	RegisterMeasurement("boolean_fields_abstract", analysis.BooleanFieldsAbstract)
	RegisterMeasurement("boolean_fields_title", analysis.BooleanFieldsTitle)
	RegisterMeasurement("boolean_fields_mesh", analysis.BooleanFieldsMeSH)
	RegisterMeasurement("boolean_fields_other", analysis.BooleanFieldsOther)
	RegisterMeasurement("boolean_and_count", analysis.BooleanAndCount)
	RegisterMeasurement("boolean_or_count", analysis.BooleanOrCount)
	RegisterMeasurement("boolean_not_count", analysis.BooleanNotCount)
	RegisterMeasurement("mesh_keywords", analysis.MeshKeywordCount)
	RegisterMeasurement("mesh_exploded", analysis.MeshExplodedCount)
	RegisterMeasurement("mesh_non_exploded", analysis.MeshNonExplodedCount)
	RegisterMeasurement("mesh_avg_depth", analysis.MeshAvgDepth)
	RegisterMeasurement("mesh_max_depth", analysis.MeshMaxDepth)

	// Evaluations measurements.
	RegisterEvaluator("precision", eval.PrecisionEvaluator)
	RegisterEvaluator("recall", eval.RecallEvaluator)
	RegisterEvaluator("num_rel", eval.NumRel)
	RegisterEvaluator("num_ret", eval.NumRet)
	RegisterEvaluator("num_rel_ret", eval.NumRelRet)
	RegisterEvaluator("f05_measure", eval.F05Measure)
	RegisterEvaluator("f1_measure", eval.F1Measure)
	RegisterEvaluator("f3_measure", eval.F3Measure)
	RegisterEvaluator("wss", eval.NewWSSEvaluator(0)) // The collection size is configured later.
	RegisterEvaluator("residual_precision", eval.NewResidualEvaluator(eval.PrecisionEvaluator))
	RegisterEvaluator("residual_recall", eval.NewResidualEvaluator(eval.RecallEvaluator))
	RegisterEvaluator("residual_f05_measure", eval.NewResidualEvaluator(eval.F05Measure))
	RegisterEvaluator("residual_f1_measure", eval.NewResidualEvaluator(eval.F1Measure))
	RegisterEvaluator("residual_f3_measure", eval.NewResidualEvaluator(eval.F3Measure))
	RegisterEvaluator("residual_wss", eval.NewResidualEvaluator(eval.NewWSSEvaluator(0))) // The collection size is configured later.
	RegisterEvaluator("mle_precision", eval.NewMaximumLikelihoodEvaluator(eval.PrecisionEvaluator))
	RegisterEvaluator("mle_recall", eval.NewMaximumLikelihoodEvaluator(eval.RecallEvaluator))
	RegisterEvaluator("mle_f05_measure", eval.NewMaximumLikelihoodEvaluator(eval.F05Measure))
	RegisterEvaluator("mle_f1_measure", eval.NewMaximumLikelihoodEvaluator(eval.F1Measure))
	RegisterEvaluator("mle_f3_measure", eval.NewMaximumLikelihoodEvaluator(eval.F3Measure))
	RegisterEvaluator("mle_wss", eval.NewMaximumLikelihoodEvaluator(eval.NewWSSEvaluator(0))) // The collection size is configured later.

	// Output formats.
	RegisterMeasurementFormatter("json", output.JsonMeasurementFormatter)
	RegisterMeasurementFormatter("csv", output.CsvMeasurementFormatter)
	RegisterEvaluationFormatter("json", output.JsonEvaluationFormatter)

	// Query Rewrite transformations.
	RegisterRewriteTransformation("logical_operator_replacement", learning.NewLogicalOperatorTransformer())
	RegisterRewriteTransformation("adj_range", learning.NewAdjacencyRangeTransformer())
	RegisterRewriteTransformation("mesh_explosion", learning.NewMeSHExplosionTransformer())
	RegisterRewriteTransformation("mesh_parent", learning.NewMeshParentTransformer())
	RegisterRewriteTransformation("field_restrictions", learning.NewFieldRestrictionsTransformer())
	RegisterRewriteTransformation("adj_replacement", learning.NewAdjacencyReplacementTransformer())
	RegisterRewriteTransformation("clause_removal", learning.NewClauseRemovalTransformer())
	err := RegisterCui2VecTransformation(dsl)
	if err != nil {
		return err
	}

	// Machine learning models.
	switch m := dsl.Learning.Model; m {
	// For the case of query chains, we need to also configure the candidate selector.
	case "query_chain":
		var model *learning.QueryChain
		var (
			depth int
			err   error
		)
		depth = 5
		if v, ok := dsl.Learning.Options["depth"]; ok {
			depth, err = strconv.Atoi(v)
			if err != nil {
				return err
			}
		}
		switch cs := dsl.Learning.Options["candidate_selector"]; cs {
		case "ltr_quickrank":
			if dsl.Learning.Train != nil {
				model = learning.NewQuickRankQueryChain(dsl.Learning.Options["binary"], dsl.Learning.Train, learning.QuickRankCandidateSelectorMaxDepth(depth))
			} else {
				model = learning.NewQuickRankQueryChain(dsl.Learning.Options["binary"], dsl.Learning.Test, learning.QuickRankCandidateSelectorMaxDepth(depth), learning.QuickRankCandidateSelectorStatisticsSource(statisticSourceMapping[dsl.Statistic.Source]))
			}
		case "reinforcement":
			model = learning.NewReinforcementQueryChain()
		case "nearest":
			if dsl.Learning.Train != nil {
				modelName := dsl.Learning.Options["model_name"]
				model = learning.NewNearestNeighbourQueryChain(learning.NearestNeighbourModelName(modelName), learning.NearestNeighbourDepth(depth))
			} else {
				modelName := dsl.Learning.Options["model_name"]
				model = learning.NewNearestNeighbourQueryChain(learning.NearestNeighbourLoadModel(modelName), learning.NearestNeighbourDepth(depth), learning.NearestNeighbourStatisticsSource(statisticSourceMapping[dsl.Statistic.Source]))
			}
		case "oracle":
			b, err := ioutil.ReadFile(dsl.Output.Evaluations.Qrels)
			if err != nil {
				return err
			}
			qrels, err := trecresults.QrelsFromReader(bytes.NewReader(b))
			if err != nil {
				return err
			}
			model = learning.NewRankOracleCandidateSelector(statisticSourceMapping[dsl.Statistic.Source], qrels, evaluationMapping[dsl.Learning.Options["measure"]], depth)
		}
		if v, ok := dsl.Learning.Options["transformed_output"]; ok {
			model.TransformedOutput = v
		}

		if v, ok := dsl.Learning.Options["features"]; ok {
			fmt.Println("loading features")
			f, err := os.Open(v)
			if err != nil {
				return err
			}
			model.LearntFeatures, err = learning.LoadFeatures(f)
			if err != nil {
				return err
			}
			fmt.Printf("loaded %d features\n", len(model.LearntFeatures))
		}

		RegisterModel(m, model)

	default:
		if len(dsl.Learning.Model) > 0 {
			return errors.New(fmt.Sprintf("could not load model of type %s", m))
		}
	}

	return nil
}
