package main

import (
	"github.com/hscells/groove/analysis"
	"github.com/hscells/groove/output"
	"github.com/hscells/groove/query"
	"github.com/hscells/groove/stats"
)

var (
	QuerySourceMapping     = map[string]query.QueriesSource{}
	StatisticSourceMapping = map[string]stats.StatisticsSource{}
	MeasurementMapping     = map[string]analysis.Measurement{}
	OutputMapping          = map[string]output.Formatter{}
)

// RegisterQuerySource registers a query source.
func RegisterQuerySource(name string, source query.QueriesSource) {
	QuerySourceMapping[name] = source
}

// RegisterStatisticSource registers a statistic source.
func RegisterStatisticSource(name string, source stats.StatisticsSource) {
	StatisticSourceMapping[name] = source
}

// RegisterMeasurement registers a measurement.
func RegisterMeasurement(name string, measurement analysis.Measurement) {
	MeasurementMapping[name] = measurement
}

// RegisterOutput registers an output formatter.
func RegisterOutput(name string, formatter output.Formatter) {
	OutputMapping[name] = formatter
}
