package main

import (
	"github.com/hscells/groove/analysis"
	"github.com/hscells/groove/output"
	"github.com/hscells/groove/query"
	"github.com/hscells/groove/stats"
)

var (
	querySourceMapping     = map[string]query.QueriesSource{}
	statisticSourceMapping = map[string]stats.StatisticsSource{}
	measurementMapping     = map[string]analysis.Measurement{}
	outputMapping          = map[string]output.Formatter{}
)

// RegisterQuerySource registers a query source.
func RegisterQuerySource(name string, source query.QueriesSource) {
	querySourceMapping[name] = source
}

// RegisterStatisticSource registers a statistic source.
func RegisterStatisticSource(name string, source stats.StatisticsSource) {
	statisticSourceMapping[name] = source
}

// RegisterMeasurement registers a measurement.
func RegisterMeasurement(name string, measurement analysis.Measurement) {
	measurementMapping[name] = measurement
}

// RegisterOutput registers an output formatter.
func RegisterOutput(name string, formatter output.Formatter) {
	outputMapping[name] = formatter
}
