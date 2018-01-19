package main

// Pipeline is a representation of the DSL.
type Pipeline struct {
	Query           PipelineQuery          `json:"query"`
	Statistic       PipelineStatistic      `json:"statistic"`
	Preprocess      []string               `json:"preprocess"`
	Measurements    []string               `json:"measurements"`
	Evaluations     []PipelineEvaluation   `json:"evaluation"`
	Transformations PipelineTransformation `json:"transformations"`
	Rewrite         PipelineRewrite        `json:"rewrite"`
	Output          PipelineOutput         `json:"output"`
}

// PipelineRewrite represents a rewrite of queries.
type PipelineRewrite struct {
	Transformations []string              `json:"transformations"`
	Chain           string                `json:"chain"`
	SVM             PipelineQueryChainSVM `json:"svm"`
}

// PipelineQueryChainSVM represents a query chain that uses an SVM.
type PipelineQueryChainSVM struct {
	Features      string `json:"features"`
	Model         string `json:"model"`
	ShouldTrain   bool   `json:"train?"`
	ShouldExtract bool   `json:"extract?"`
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

// EvaluationOutputFormat represents how evaluations should be output.
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
