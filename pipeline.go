package boogie

import (
	"bytes"
	"fmt"
	"github.com/hscells/groove"
	"github.com/hscells/groove/analysis"
	"github.com/hscells/groove/eval"
	"github.com/hscells/groove/learning"
	"github.com/hscells/groove/output"
	"github.com/hscells/groove/preprocess"
	"github.com/hscells/trecresults"
	"io/ioutil"
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
				if _, ok := m.(eval.WorkSavedOverSampling); ok {
					mm := m.(eval.WorkSavedOverSampling)
					mm.N, err = g.StatisticsSource.CollectionSize()
					if err != nil {
						return g, err
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
