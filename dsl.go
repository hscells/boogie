package boogie

// Pipeline is a representation of the DSL.
type Pipeline struct {
	Query           PipelineQuery          `json:"query"`
	Statistic       PipelineStatistic      `json:"statistic"`
	MetaWrap        string                 `json:"metawrap"`
	CUIMapping      string                 `json:"cui_mapping"`
	CUI2vec         string                 `json:"cui2vec"`
	Preprocess      []string               `json:"preprocess"`
	Measurements    []string               `json:"measurements"`
	Evaluations     []string               `json:"evaluation"`
	Transformations PipelineTransformation `json:"transformations"`
	Rewrite         PipelineRewrite        `json:"rewrite"`
	Output          PipelineOutput         `json:"output"`
	Cache           []PipelineCache        `json:"cache"`
}

// PipelineMetaWrap configures the metawrap server.
type PipelineMetaWrap struct {
	URL string `json:"url"`
}

// PipelineCUIMapping configures the CUI mapping service.
type PipelineCUIMapping struct {
	Path string `json:"path"`
}

// PipelineCache configures caching.
type PipelineCache struct {
	Type    string                 `json:"type"`
	Options map[string]interface{} `json:"options"`
}

// PipelineRewrite represents a rewrite of queries.
type PipelineRewrite struct {
	Transformations []string          `json:"transformations"`
	Chain           string            `json:"chain"`
	Options         map[string]string `json:"options"`
}

// PipelineQueryChainSVM represents a query chain that uses an SVM.
type PipelineQueryChainSVM struct {
	Model string `json:"model"`
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
