package boogie

import (
	"bytes"
	"fmt"
	"github.com/hscells/groove/analysis"
	"github.com/hscells/groove/eval"
	"github.com/hscells/groove/learning"
	"github.com/hscells/groove/output"
	"github.com/hscells/groove/pipeline"
	"github.com/hscells/groove/preprocess"
	"github.com/hscells/trecresults"
	"io/ioutil"
)

// CreatePipeline creates the main groove pipeline.
func CreatePipeline(dsl Pipeline) (pipeline.GroovePipeline, error) {
	// Register the sources used in the groove pipeline.
	err := RegisterSources(dsl)
	if err != nil {
		return pipeline.GroovePipeline{}, err
	}

	// Create a groove pipeline from the boogie dsl.
	g := pipeline.GroovePipeline{}

	if len(dsl.Query.Format) > 0 {
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

	if len(dsl.Measurements) > 0 && len(dsl.Output.Measurements) == 0 {
		return g, fmt.Errorf("at least one output format must be supplied when using analysis measurements")
	}

	if len(dsl.Output.Measurements) > 0 && len(dsl.Measurements) == 0 {
		return g, fmt.Errorf("at least one analysis measurement must be supplied for the output formats")
	}

	if len(dsl.Evaluations) > 0 && len(dsl.Output.Evaluations.Measurements) == 0 {
		return g, fmt.Errorf("at least one output format must be supplied when using evaluation measurements")
	}

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

	//g.QueryChain
	if len(dsl.Rewrite) > 0 {
		var transformations []learning.Transformation
		for _, transformation := range dsl.Rewrite {
			if t, ok := rewriteTransformationMapping[transformation]; ok {
				transformations = append(transformations, t)
			} else {
				return g, fmt.Errorf("%v is not a known rewrite transformation", transformation)
			}
		}
	}

	g.Transformations.Output = dsl.Transformations.Output
	g.OutputTrec.Path = dsl.Output.Trec.Output
	return g, nil
}
