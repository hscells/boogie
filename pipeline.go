package boogie

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/hscells/cui2vec"
	"github.com/hscells/groove"
	"github.com/hscells/groove/analysis"
	"github.com/hscells/groove/eval"
	"github.com/hscells/groove/learning"
	"github.com/hscells/groove/output"
	"github.com/hscells/groove/preprocess"
	"github.com/hscells/trecresults"
	"io/ioutil"
	"os"
)

// CreatePipeline creates the main groove pipeline.
func CreatePipeline(dsl Pipeline) (groove.Pipeline, error) {
	// Register the sources used in the groove pipeline.
	err := RegisterSources(dsl)
	if err != nil {
		return groove.Pipeline{}, err
	}

	// Create a groove pipeline from the boogie dsl.
	g := groove.Pipeline{}
	g.QueryPath = dsl.Query.Path

	if len(dsl.Query.Path) > 0 {
		if s, ok := querySourceMapping[dsl.Query.Format]; ok {
			g.QueriesSource = s
		} else {
			return g, fmt.Errorf("%v is not a known query source", dsl.Query.Format)
		}
	}

	if len(dsl.Statistic.Source) > 0 {
		if s, ok := statisticSourceMapping[dsl.Statistic.Source]; ok {
			g.StatisticsSource = s
		} else {
			return g, fmt.Errorf("%v is not a known statistics source", dsl.Statistic.Source)
		}
	}

	if len(dsl.Statistic.Source) == 0 && len(dsl.Measurements) > 0 {
		return g, fmt.Errorf("a statistic source is required for measurements")
	}

	//if len(dsl.Measurements) > 0 && len(dsl.Output.Measurements) == 0  {
	//	return g, fmt.Errorf("at least one output format must be supplied when using analysis measurements")
	//}

	if len(dsl.Output.Measurements) > 0 && len(dsl.Measurements) == 0 {
		return g, fmt.Errorf("at least one analysis measurement must be supplied for the output formats")
	}

	//if len(dsl.Evaluations) > 0 && len(dsl.Output.Evaluations.Measurements) == 0 {
	//	return g, fmt.Errorf("at least one output format must be supplied when using evaluation measurements")
	//}

	if len(dsl.Output.Evaluations.Measurements) > 0 && len(dsl.Evaluations) == 0 {
		return g, fmt.Errorf("at least one evaluation measurement must be supplied for the output formats")
	}

	g.Measurements = []analysis.Measurement{}
	for _, measurementName := range dsl.Measurements {
		if m, ok := measurementMapping[measurementName]; ok {
			g.Measurements = append(g.Measurements, m)
		} else {
			return g, fmt.Errorf("%v is not a known measurement", measurementName)
		}
	}

	g.Evaluations = []eval.Evaluator{}
	for _, measurement := range dsl.Evaluations {
		if m, ok := evaluationMapping[measurement]; ok {
			// Here we configure wss directly with the N component (collection size).
			if measurement == "wss" {
				N, err := g.StatisticsSource.CollectionSize()
				if err != nil {
					return g, err
				}
				if _, ok := m.(eval.WorkSavedOverSampling); ok {
					mm := m.(eval.WorkSavedOverSampling)
					mm.N = N
					m = mm
				}
			} else if measurement == "residual_wss" {
				N, err := g.StatisticsSource.CollectionSize()
				if err != nil {
					return g, err
				}
				if _, ok := m.(eval.ResidualEvaluator); ok {
					mm := m.(eval.ResidualEvaluator)
					if v, ok := mm.Evaluator.(eval.WorkSavedOverSampling); ok {
						v.N = N
						mm.Evaluator = v
					}
					m = mm
				}
			} else if measurement == "mle_wss" {
				N, err := g.StatisticsSource.CollectionSize()
				if err != nil {
					return g, err
				}
				if _, ok := m.(eval.MaximumLikelihoodEvaluator); ok {
					mm := m.(eval.MaximumLikelihoodEvaluator)
					if v, ok := mm.Evaluator.(eval.WorkSavedOverSampling); ok {
						v.N = N
						mm.Evaluator = v
					}
					m = mm
				}
			}
			g.Evaluations = append(g.Evaluations, m)
		} else {
			return g, fmt.Errorf("%v is not a known evaluation measurement", measurement)
		}
	}

	if len(dsl.Output.Evaluations.Qrels) > 0 {
		b, err := ioutil.ReadFile(dsl.Output.Evaluations.Qrels)
		if err != nil {
			return g, err
		}
		qrels, err := trecresults.QrelsFromReader(bytes.NewReader(b))
		if err != nil {
			return g, err
		}
		g.EvaluationFormatters.EvaluationQrels = qrels
	}

	g.MeasurementFormatters = []output.MeasurementFormatter{}
	for _, formatter := range dsl.Output.Measurements {
		if o, ok := measurementFormatters[formatter.Format]; ok {
			g.MeasurementFormatters = append(g.MeasurementFormatters, o)
		} else {
			return g, fmt.Errorf("%v is not a known measurement output format", formatter.Format)
		}
	}

	g.EvaluationFormatters.EvaluationFormatters = []output.EvaluationFormatter{}
	for _, formatter := range dsl.Output.Evaluations.Measurements {
		if o, ok := evaluationFormatters[formatter.Format]; ok {
			g.EvaluationFormatters.EvaluationFormatters = append(g.EvaluationFormatters.EvaluationFormatters, o)
		} else {
			return g, fmt.Errorf("%v is not a known evaluation output format", formatter.Format)
		}
	}

	g.Preprocess = []preprocess.QueryProcessor{}
	for _, p := range dsl.Preprocess {
		if processor, ok := preprocessorMapping[p]; ok {
			g.Preprocess = append(g.Preprocess, processor)
		} else {
			return g, fmt.Errorf("%v is not a known preprocessor", p)
		}
	}

	g.Transformations = preprocess.QueryTransformations{}
	for _, t := range dsl.Transformations.Operations {
		if transformation, ok := transformationMappingBoolean[t]; ok {
			g.Transformations.BooleanTransformations = append(g.Transformations.BooleanTransformations, transformation)
		} else if transformation, ok := transformationMappingElasticsearch[t]; ok {
			g.Transformations.ElasticsearchTransformations = append(g.Transformations.ElasticsearchTransformations, transformation)
		} else {
			return g, fmt.Errorf("%v is not a known preprocessing transformation", t)
		}
	}

	// Configure the transformations that can be applied in the context of a query chain model.
	var transformations []learning.Transformation
	if len(dsl.Rewrite) > 0 && len(dsl.Learning.Model) > 0 {
		for _, transformation := range dsl.Rewrite {
			if t, ok := rewriteTransformationMapping[transformation]; ok {
				transformations = append(transformations, t)
			} else {
				return g, fmt.Errorf("%v is not a known rewrite transformation", transformation)
			}
		}
	}

	// Configure the learning model to use.
	if len(dsl.Learning.Model) > 0 {
		if m, ok := modelMapping[dsl.Learning.Model]; ok {
			if dsl.Learning.Train != nil {
				g.ModelConfiguration.Train = true
			}
			if dsl.Learning.Test != nil {
				g.ModelConfiguration.Test = true
			}
			if dsl.Learning.Generate != nil {
				g.ModelConfiguration.Generate = true
			}

			g.Model = m
			switch m := g.Model.(type) {
			case *learning.QueryChain:
				m.QrelsFile = g.EvaluationFormatters.EvaluationQrels
				m.Evaluators = g.Evaluations
				m.Measurements = g.Measurements
				m.Transformations = transformations
				m.StatisticsSource = g.StatisticsSource
				m.QrelsFile = g.EvaluationFormatters.EvaluationQrels
				if g.ModelConfiguration.Generate {
					m.GenerationFile = dsl.Learning.Generate["output"].(string)
					if s, ok := dsl.Learning.Generate["sampler"].(string); ok {
						var (
							n       int
							delta   float64
							measure string
						)
						if v, ok := dsl.Learning.Generate["n"]; ok {
							n = int(v.(float64))
						}
						if v, ok := dsl.Learning.Generate["delta"].(float64); ok {
							delta = v
						}
						if v, ok := dsl.Learning.Generate["measure"].(string); ok {
							measure = v
						}
						if n == 0 || delta == 0 {
							return groove.Pipeline{}, errors.New("neither n or delta are configured for sampling (cannot be 0 values)")
						}
						switch s {
						case "greedy":
							if len(measure) == 0 {
								return groove.Pipeline{}, errors.New("mis-configured measure for greedy sampler")
							}
							var (
								e        eval.Evaluator
								strategy learning.GreedyStrategy
								scores   map[string]float64
							)

							// Configure the evaluation measure used in sampling.
							if m, ok := evaluationMapping[measure]; ok {
								e = m
							} else {
								return groove.Pipeline{}, fmt.Errorf("%s is not a valid evaluation measure for sampling", measure)
							}

							// Configure loading of the scores for sampling.
							if v, ok := dsl.Learning.Generate["scores"]; ok {
								path := v.(string)
								f, err := os.OpenFile(path, os.O_RDONLY, os.ModePerm)
								if err != nil {
									return groove.Pipeline{}, err
								}
								b, err := ioutil.ReadAll(f)
								if err != nil {
									return groove.Pipeline{}, err
								}
								json.Unmarshal(b, &scores)
							}

							// Configure the sampling strategy.
							if v, ok := dsl.Learning.Generate["strategy"]; ok {
								switch v.(string) {
								case "naive":
									strategy = learning.RankedGreedyStrategy
								case "diversified":
									strategy = learning.MaximalMarginalRelevanceGreedyStrategy(scores, 0.3, cui2vec.Cosine)
								}
							} else {
								return groove.Pipeline{}, fmt.Errorf("unknown greedy sampling strategy %v", v)
							}

							m.Sampler = learning.NewGreedySampler(n, delta, e, g.EvaluationFormatters.EvaluationQrels, g.QueryCache, g.StatisticsSource, strategy)
							break
						case "evaluation":
							if len(measure) == 0 {
								return groove.Pipeline{}, errors.New("mis-configured measure for evaluation sampler")
							}
							var (
								e        eval.Evaluator
								strategy learning.ScoredStrategy
								scores   map[string]float64
							)

							// Configure the evaluation measure used in sampling.
							if m, ok := evaluationMapping[measure]; ok {
								e = m
							} else {
								return groove.Pipeline{}, fmt.Errorf("%s is not a valid evaluation measure for sampling", measure)
							}

							// Configure the sampling strategy.
							if v, ok := dsl.Learning.Generate["strategy"]; ok {
								switch v.(string) {
								case "positive_biased":
									strategy = learning.PositiveBiasScoredStrategy
								case "negative_biased":
									strategy = learning.NegativeBiasScoredStrategy
								case "balanced":
									strategy = learning.BalancedScoredStrategy
								case "stratified":
									strategy = learning.StratifiedScoredStrategy
								case "diversified":
									strategy = learning.MaximalMarginalRelevanceScoredStrategy(0.3, cui2vec.Cosine)
								}
							} else {
								return groove.Pipeline{}, fmt.Errorf("unknown evaluation sampling strategy %v", v)
							}

							// Configure loading of the scores for sampling.
							if v, ok := dsl.Learning.Generate["scores"]; ok {
								path := v.(string)
								f, err := os.OpenFile(path, os.O_RDONLY, os.ModePerm)
								if err != nil {
									return groove.Pipeline{}, err
								}
								b, err := ioutil.ReadAll(f)
								if err != nil {
									return groove.Pipeline{}, err
								}
								json.Unmarshal(b, &scores)
							}

							m.Sampler = learning.NewEvaluationSampler(n, delta, e, g.EvaluationFormatters.EvaluationQrels, g.QueryCache, g.StatisticsSource, scores, strategy)
							break
						case "transformation":
							var strategy learning.TransformationStrategy

							// Configure the sampling strategy.
							if v, ok := dsl.Learning.Generate["strategy"]; ok {
								switch v.(string) {
								case "stratified":
									strategy = learning.StratifiedTransformationStrategy
								case "balanced":
									strategy = learning.BalancedTransformationStrategy
								}
							} else {
								return groove.Pipeline{}, fmt.Errorf("unknown transformation sampling strategy %v", v)
							}

							m.Sampler = learning.NewTransformationSampler(n, delta, strategy)
							break
						case "random":
							m.Sampler = learning.NewRandomSampler(n, delta)
						case "cluster":
							var k int
							// Configure the evaluation measure used in sampling.
							if v, ok := dsl.Learning.Generate["k"]; ok {
								k, ok = v.(int)
								if !ok {
									return groove.Pipeline{}, fmt.Errorf("%s is not an integer for k", v)
								}
							} else {
								return groove.Pipeline{}, fmt.Errorf("%s is not a valid value for k", v)
							}

							m.Sampler = learning.NewClusterSampler(n, delta, k)
						default:
							return groove.Pipeline{}, fmt.Errorf("%s is not a valid sampler", s)
						}
					} else {
						return groove.Pipeline{}, fmt.Errorf("ensure that a sampler is configured when generating data")
					}
				}
			default:
				return g, fmt.Errorf("unable to properly configure the learning model %s, see pipeline.go for more information", dsl.Learning.Model)
			}
		} else {
			return g, fmt.Errorf("%s is not a known learning model", dsl.Learning.Model)
		}
	}

	g.Transformations.Output = dsl.Transformations.Output
	g.OutputTrec.Path = dsl.Output.Trec.Output
	return g, nil
}
