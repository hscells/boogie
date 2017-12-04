package main

// Pipeline is a representation of the DSL.
type Pipeline struct {
	Query           PipelineQuery          `json:"query"`
	Statistic       PipelineStatistic      `json:"statistic"`
	Preprocess      []string               `json:"preprocess"`
	Measurements    []string               `json:"measurements"`
	Evaluations     []PipelineEvaluation   `json:"evaluation"`
	Transformations PipelineTransformation `json:"transformations"`
	Output          PipelineOutput         `json:"output"`
}

// PipelineQuery represents a query source in the DSL.
type PipelineQuery struct {
	Format        string                 `json:"format"`
	Options       map[string]interface{} `json:"options"`
	SearchOptions map[string]interface{} `json:"search"`
}

// PipelineStatistic represents a statistic source in the DSL.
type PipelineStatistic struct {
	Source  string                 `json:"source"`
	Options map[string]interface{} `json:"options"`
}

// PipelineEvaluation represents what evaluation measures should be computed, and any parameters configuration for them.
type PipelineEvaluation struct {
	Evaluation string `json:"evaluate"`
}

// PipelineOutput represents an output formatter in the DSL.
type PipelineOutput struct {
	Measurements []MeasurementOutput `json:"measurements"`
	Trec         TrecOutput          `json:"trec_results"`
	Evaluations  EvaluationOutput    `json:"evaluations"`
}

// MeasurementOutput represents an output format for measurements.
type MeasurementOutput struct {
	Format   string `json:"format"`
	Filename string `json:"filename"`
}

// MeasurementOutput represents an output format for measurements.
type EvaluationOutput struct {
	Qrels        string                   `json:"qrels"`
	Measurements []EvaluationOutputFormat `json:"formats"`
}

type EvaluationOutputFormat struct {
	Format   string `json:"format"`
	Filename string `json:"filename"`
}

// TrecOutput represents an output for trec files.
type TrecOutput struct {
	Output string `json:"output"`
}

// PipelineTransformation represents an set of transformation operations in the DSL.
type PipelineTransformation struct {
	Output     string   `json:"output"`
	Operations []string `json:"operations"`
}
