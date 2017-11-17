package main

// Pipeline is a representation of the DSL.
type Pipeline struct {
	Query        PipelineQuery     `json:"query"`
	Statistic    PipelineStatistic `json:"statistic"`
	Measurements []string          `json:"measurements"`
	Output       []PipelineOutput  `json:"output"`
}

// PipelineQuery represents a query source in the DSL.
type PipelineQuery struct {
	Source string `json:"source"`
}

// PipelineStatistic represents a statistic source in the DSL.
type PipelineStatistic struct {
	Source string `json:"source"`
}

// PipelineOutput represents an output formatter in the DSL.
type PipelineOutput struct {
	Format   string `json:"format"`
	Filename string `json:"filename"`
}
