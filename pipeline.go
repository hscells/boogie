package boogie

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/hscells/cqr"
	"github.com/hscells/cui2vec"
	"github.com/hscells/groove"
	"github.com/hscells/groove/analysis"
	"github.com/hscells/groove/eval"
	"github.com/hscells/groove/formulation"
	"github.com/hscells/groove/learning"
	"github.com/hscells/groove/output"
	"github.com/hscells/groove/preprocess"
	"github.com/hscells/groove/query"
	"github.com/hscells/groove/stats"
	"github.com/hscells/guru"
	"github.com/hscells/metawrap"
	"github.com/hscells/trecresults"
	"io/ioutil"
	"os"
	"path"
	"strconv"
)

// CreatePipeline creates the main groove pipeline.
func CreatePipeline(dsl Pipeline) (groove.Pipeline, error) {
	// Register the sources used in the groove pipeline.
	err := RegisterSources(dsl)
	if err != nil {
		return groove.Pipeline{}, err
	}

	eval.RelevanceGrade = dsl.Output.Evaluations.RelevanceGrade

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
				if g.ModelConfiguration.Generate {
					m.GenerationFile = dsl.Learning.Generate["output"].(string)

					traversal := dsl.Learning.Generate["traversal"].(string)
					if traversal == "depth_first" {
						var (
							measure string
							sampler learning.DepthFirstSamplingCriteria
							budget  int
						)
						if v, ok := dsl.Learning.Generate["budget"]; ok {
							budget = int(v.(float64))
						}

						if s, ok := dsl.Learning.Generate["sampler"].(string); ok {
							if v, ok := dsl.Learning.Generate["measure"].(string); ok {
								measure = v
							}
							switch s {
							case "evaluation":
								var (
									e      eval.Evaluator
									scores map[string]map[string]float64
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
									err = json.Unmarshal(b, &scores)
									if err != nil {
										return groove.Pipeline{}, err
									}
								} else {
									return groove.Pipeline{}, errors.New("no scores parameter defined")
								}

								// Configure the sampling strategy.
								if v, ok := dsl.Learning.Generate["strategy"]; ok {
									switch v.(string) {
									case "positive":
										sampler = learning.PositiveBiasedEvaluationSamplingCriteria(e, scores, m)
									case "negative":
										sampler = learning.NegativeBiasedEvaluationSamplingCriteria(e, scores, m)
									case "balanced":
										sampler = learning.BalancedEvaluationSamplingCriteria(e, scores, m)
									}
								} else {
									return groove.Pipeline{}, fmt.Errorf("unknown greedy sampling strategy %v", v)
								}
							case "transformation":
								// Configure the sampling strategy.
								if v, ok := dsl.Learning.Generate["strategy"]; ok {
									switch v.(string) {
									case "balanced":
										sampler = learning.BalancedTransformationSamplingCriteria(learning.ChainFeatures)
									case "biased":
										sampler = learning.BiasedTransformationSamplingCriteria()
									}
								} else {
									return groove.Pipeline{}, fmt.Errorf("unknown greedy sampling strategy %v", v)
								}
							case "random":
								sampler = learning.ProbabilisticSamplingCriteria(0.65)
							}
						} else {
							return groove.Pipeline{}, fmt.Errorf("ensure that a sampler is configured when generating data")
						}
						m.GenerationExplorer = learning.NewDepthFirstExplorer(m, sampler, budget)
					} else if traversal == "breadth_first" {
						var (
							n       int
							delta   float64
							measure string
							sampler learning.Sampler
						)
						if s, ok := dsl.Learning.Generate["sampler"].(string); ok {
							if v, ok := dsl.Learning.Generate["n"]; ok {
								n = int(v.(float64))
							}
							if v, ok := dsl.Learning.Generate["delta"].(float64); ok {
								delta = v
							}
							if v, ok := dsl.Learning.Generate["measure"].(string); ok {
								measure = v
							}
							switch s {
							case "greedy":
								if len(measure) == 0 {
									return groove.Pipeline{}, errors.New("mis-configured measure for greedy sampler")
								}
								var (
									e        eval.Evaluator
									strategy learning.GreedyStrategy
									scores   map[string]map[string]float64
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
									err = json.Unmarshal(b, &scores)
									if err != nil {
										return groove.Pipeline{}, err
									}
								} else {
									return groove.Pipeline{}, errors.New("no scores parameter defined")
								}

								// Configure the sampling strategy.
								if v, ok := dsl.Learning.Generate["strategy"]; ok {
									switch v.(string) {
									case "naive":
										strategy = learning.RankedGreedyStrategy
									case "diversified":
										strategy = learning.MaximalMarginalRelevanceGreedyStrategy(scores, 0.3, cui2vec.Cosine, e)
									}
								} else {
									return groove.Pipeline{}, fmt.Errorf("unknown greedy sampling strategy %v", v)
								}

								sampler = learning.NewGreedySampler(n, delta, e, m, strategy)
								break
							case "evaluation":
								if len(measure) == 0 {
									return groove.Pipeline{}, errors.New("mis-configured measure for evaluation sampler")
								}
								var (
									e        eval.Evaluator
									strategy learning.ScoredStrategy
									scores   map[string]map[string]float64
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
									err = json.Unmarshal(b, &scores)
									if err != nil {
										return groove.Pipeline{}, err
									}
								} else {
									return groove.Pipeline{}, errors.New("no scores parameter defined")
								}

								sampler = learning.NewEvaluationSampler(n, delta, e, m, scores, strategy)
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

								sampler = learning.NewTransformationSampler(n, delta, strategy)
								break
							case "random":
								sampler = learning.NewRandomSampler(n, delta)
							case "cluster":
								var k int
								// Configure the evaluation measure used in sampling.
								if v, ok := dsl.Learning.Generate["k"].(float64); ok {
									k = int(k)
								} else {
									return groove.Pipeline{}, fmt.Errorf("%v is not a valid value for k", v)
								}

								sampler = learning.NewClusterSampler(n, delta, k)
							default:
								return groove.Pipeline{}, fmt.Errorf("%s is not a valid sampler", s)
							}
						} else {
							return groove.Pipeline{}, fmt.Errorf("ensure that a sampler is configured when generating data")
						}

						m.GenerationExplorer = learning.NewBreadthFirstExplorer(m, sampler, learning.DepthStoppingCondition(5))
					}
				}
			default:
				return g, fmt.Errorf("unable to properly configure the learning model %s, see pipeline.go for more information", dsl.Learning.Model)
			}
		} else {
			return g, fmt.Errorf("%s is not a known learning model", dsl.Learning.Model)
		}
	}

	// Configure the query formulation method.
	if len(dsl.Formulation.Method) > 0 {
		switch dsl.Formulation.Method {
		case "conceptual":
			title := dsl.Formulation.Options["title"]
			topic := dsl.Formulation.Options["topic"]

			var (
				composer   formulation.LogicComposer
				extractor  formulation.EntityExtractor
				expander   formulation.EntityExpander
				mapper     formulation.KeywordMapper
				processing []formulation.PostProcess
			)

			switch dsl.Formulation.Options["logic_composer"] {
			case "nlp":
				composer = formulation.NewNLPLogicComposer(dsl.Formulation.Options["logic_composer.classpath"])
			case "manual":
				composer = formulation.NewManualLogicComposer(dsl.Formulation.Options["logic_composer.output_path"], topic)
			}

			switch dsl.Formulation.Options["entity_extractor"] {
			case "metamap":
				extractor = formulation.NewMetaMapEntityExtractor(metawrap.HTTPClient{URL: dsl.Formulation.Options["metamap_url"]})
			}

			switch dsl.Formulation.Options["entity_expander"] {
			case "cui2vec":
				f, err := os.OpenFile(dsl.Formulation.Options["entity_expander.cui2vec_precomputed_embeddings"], os.O_RDONLY, 0664)
				if err != nil {
					return g, err
				}
				e, err := cui2vec.NewPrecomputedEmbeddings(f)
				if err != nil {
					return g, err
				}
				err = f.Close()
				if err != nil {
					return g, err
				}
				expander = formulation.NewCui2VecEntityExpander(*e)
			case "medgen":
				e, err := NewEntrezStatisticsSource(dsl.Statistic.Options, stats.EntrezDb("medgen"))
				if err != nil {
					return g, err
				}
				expander = formulation.NewMedGenExpander(e)
			default:
				expander = nil
			}

			var pmids []int
			switch dsl.Formulation.Options["relevance_feedback"] {
			case "rf":
				f, err := os.OpenFile(dsl.Formulation.Options["relevance_feedback.feedback"], os.O_RDONLY, 0664)
				if err != nil {
					return groove.Pipeline{}, err
				}
				s := bufio.NewScanner(f)
				for s.Scan() {
					line := s.Text()
					x, err := strconv.Atoi(line)
					if err != nil {
						return groove.Pipeline{}, err
					}
					pmids = append(pmids, x)
				}

			}

			switch dsl.Formulation.Options["keyword_mapper"] {
			case "metamap":
				var m formulation.MetaMapMapper
				switch dsl.Formulation.Options["keyword_mapper.mapper"] {
				case "matched":
					m = formulation.Matched()
				case "preferred":
					client, err := guru.NewUMLSClient(dsl.Formulation.Options["keyword_mapper.mapper.umls_username"], dsl.Formulation.Options["keyword_mapper.mapper.umls_password"])
					if err != nil {
						return g, err
					}
					m = formulation.Preferred(client)
				case "frequent":
					c, err := cui2vec.LoadCUIMapping(dsl.Formulation.Options["keyword_mapper.mapper.cui2vec_frequent_mapping"])
					if err != nil {
						return g, err
					}
					m = formulation.Frequent(c)
				case "alias":
					c, err := cui2vec.LoadCUIAliasMapping(dsl.Formulation.Options["keyword_mapper.mapper.cui2vec_alias_mapping"])
					if err != nil {
						return g, err
					}
					m = formulation.Alias(c)
				}
				if len(dsl.Formulation.Options["keyword_mapper.add_mesh"]) > 0 {
					m = formulation.MeSHMapper(m)
				}
				mapper = formulation.NewMetaMapKeywordMapper(metawrap.HTTPClient{URL: dsl.Formulation.Options["metamap_url"]}, m)
			}

			for _, pp := range dsl.Formulation.PostProcessing {
				switch pp {
				case "stem":
					// Find the original query so as to stem it.
					queries, err := query.TARTask2QueriesSource{}.Load(dsl.Formulation.Options["post_processing.tar_topics_path"])
					if err != nil {
						return g, err
					}
					var original cqr.CommonQueryRepresentation
					for _, q := range queries {
						if q.Topic == topic {
							original = q.Query
						}
					}
					processing = append(processing, formulation.Stem(original))
				}
			}

			g.QueryFormulator = formulation.NewConceptualFormulator(title, topic, composer, extractor, expander, mapper, pmids, g.StatisticsSource.(stats.EntrezStatisticsSource), processing...)
		case "objective":
			topic := dsl.Formulation.Options["topic"]
			folder := dsl.Formulation.Options["folder"]
			pubdates := dsl.Formulation.Options["pubdates"]
			semtypes := dsl.Formulation.Options["semtypes"]
			metamap := dsl.Formulation.Options["metamap"]
			optimisation, ok := evaluationMapping[dsl.Formulation.Options["optimisation"]]
			if !ok {
				return groove.Pipeline{}, fmt.Errorf("%s is not a known evaluation measure", dsl.Formulation.Options["optimisation"])
			}

			// Find the original query so as to stem it.
			input, err := query.TARTask2QueriesSource{}.LoadSingle(path.Join(dsl.Formulation.Options["tar_topics_path"], topic))
			if err != nil {
				return g, err
			}

			var (
				processing []formulation.PostProcess
				population formulation.BackgroundCollection
				analyser   formulation.TermAnalyser
				splitter   formulation.Splitter
			)

			switch dsl.Formulation.Options["analyser"] {
			case "term_frequency":
				analyser = formulation.TermFrequencyAnalyser
			}

			switch dsl.Formulation.Options["background_collection"] {
			case "pubmed":
				population = formulation.NewPubMedSet(g.StatisticsSource.(stats.EntrezStatisticsSource))
			case "top10000":
				population, err = formulation.GetPopulationSet(g.StatisticsSource.(stats.EntrezStatisticsSource), analyser)
				if err != nil {
					return g, err
				}
			}

			switch dsl.Formulation.Options["splitter"] {
			case "random":
				splitter = formulation.RandomSplitter(1000)
			}

			for _, pp := range dsl.Formulation.PostProcessing {
				switch pp {
				case "stem":
					// Find the original query so as to stem it.
					queries, err := query.TARTask2QueriesSource{}.Load(dsl.Formulation.Options["post_processing.tar_topics_path"])
					if err != nil {
						return g, err
					}
					var original cqr.CommonQueryRepresentation
					for _, q := range queries {
						if q.Topic == topic {
							original = q.Query
						}
					}
					processing = append(processing, formulation.Stem(original))
				}
			}
			qrels := g.EvaluationFormatters.EvaluationQrels.Qrels
			g.QueryFormulator = formulation.NewObjectiveFormulator(input, g.StatisticsSource.(stats.EntrezStatisticsSource), qrels[topic], population, folder, pubdates, semtypes, metamap, optimisation,
				formulation.ObjectiveAnalyser(analyser),
				formulation.ObjectiveSplitter(splitter))
		case "dt":
			qrels := g.EvaluationFormatters.EvaluationQrels.Qrels
			topic := dsl.Formulation.Options["topic"]
			if qrels == nil {
				return g, errors.New("qrels file has not been specified for dt formulator")
			}

			var p, n []int
			for _, line := range qrels[topic] {
				id, err := strconv.Atoi(line.DocId)
				if err != nil {
					return g, err
				}
				if line.Score > eval.RelevanceGrade {
					p = append(p, id)
				} else {
					n = append(n, id)
				}
			}

			e, ok := g.StatisticsSource.(stats.EntrezStatisticsSource)
			if !ok {
				return g, errors.New("the entrez statistics source must be configured to use the dt formulator")

			}
			fmt.Println(len(p), len(n))

			pos, err := e.Fetch(p)
			if err != nil {
				return g, err
			}

			var neg guru.MedlineDocuments
			if len(n) > 0 {
				neg, err = e.Fetch(n)
				if err != nil {
					return g, err
				}
			} else {
				neg = make(guru.MedlineDocuments, 0)
			}

			fmt.Println(len(pos), len(neg))

			g.QueryFormulator, err = formulation.NewDecisionTreeFormulator(topic, pos, neg)
			if err != nil {
				return g, err
			}

		default:
			return g, fmt.Errorf("no such query formulation method: %s", dsl.Formulation.Method)
		}
	}

	g.CLF = dsl.CLFOptions

	g.Transformations.Output = dsl.Transformations.Output
	g.OutputTrec.Path = dsl.Output.Trec.Output
	return g, nil
}
