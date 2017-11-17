package main

// Pipeline is a representation of the DSL.
type Pipeline struct {
	Query        PipelineQuery     `json:"query"`
	Statistic    PipelineStatistic `json:"statistic"`
	Preprocess   []string          `json:"preprocess"`
	Measurements []string          `json:"measurements"`
	Output       []PipelineOutput  `json:"output"`
}

// PipelineQuery represents a query source in the DSL.
type PipelineQuery struct {
	Format string `json:"format"`
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
