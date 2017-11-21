package main

// Pipeline is a representation of the DSL.
type Pipeline struct {
	Query           PipelineQuery            `json:"query"`
	Statistic       PipelineStatistic        `json:"statistic"`
	Preprocess      []string                 `json:"preprocess"`
	Transformations PipelineTransformation   `json:"transformations"`
	Measurements    []string                 `json:"measurements"`
	Output          []PipelineOutput         `json:"output"`
}

// PipelineQuery represents a query source in the DSL.
type PipelineQuery struct {
	Format  string                 `json:"format"`
	Options map[string]interface{} `json:"options"`
}

// PipelineStatistic represents a statistic source in the DSL.
type PipelineStatistic struct {
	Source  string                 `json:"source"`
	Options map[string]interface{} `json:"options"`
}

// PipelineOutput represents an output formatter in the DSL.
type PipelineOutput struct {
	Format   string `json:"format"`
	Filename string `json:"filename"`
}

// PipelineTransformation represents an set of transformation operations in the DSL.
type PipelineTransformation struct {
	Output     string   `json:"output"`
	Operations []string `json:"operations"`
}
