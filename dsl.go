package boogie

// Pipeline is a representation of the DSL.
type Pipeline struct {
	Query           PipelineQuery          `json:"query"`
	Statistic       PipelineStatistic      `json:"statistic"`
	Utilities       PipelineUtilities      `json:"utilities"`
	Preprocess      []string               `json:"preprocess"`
	Measurements    []string               `json:"measurements"`
	Evaluations     []string               `json:"evaluation"`
	Transformations PipelineTransformation `json:"transformations"`
	Learning        PipelineLearning       `json:"learning"`
	Rewrite         []string               `json:"rewrite"`
	Output          PipelineOutput         `json:"output"`
	Cache           []PipelineCache        `json:"cache"`
}

// PipelineUtilities is used to reference external tools or files.
type PipelineUtilities struct {
	MetaWrap    string `json:"metawrap"`
	CUIMapping  string `json:"cui_mapping"`
	CUI2vec     string `json:"cui2vec"`
	CUI2vecSkip bool   `json:"cui2vec_skip_first"`
}

// PipelineCache configures caching.
type PipelineCache struct {
	Type    string                 `json:"type"`
	Options map[string]interface{} `json:"options"`
}

// PipelineLearning represents a configuration of a learning model.
// The model specified in `model` is configured via `options`.
//
// The train, test, validate, and generate methods can be configured
// via the options of the same names.
type PipelineLearning struct {
	Model    string            `json:"model"`
	Options  map[string]string `json:"options"`
	Train    map[string]string `json:"train"`
	Test     map[string]string `json:"test"`
	Validate map[string]string `json:"validate"`
	Generate map[string]string `json:"generate"`
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

// EvaluationOutput represents an output format for measurements.
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
