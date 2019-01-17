//+build windows

package boogie

import "github.com/hscells/groove/stats"

// NewTerrierStatisticsSource attempts to create a terrier statistics source.
func NewTerrierStatisticsSource(config map[string]interface{}) *stats.TerrierStatisticsSource {
	var propsFile string
	field := "text"

	if pf, ok := config["properties"]; ok {
		propsFile = pf.(string)
	}

	if f, ok := config["field"]; ok {
		field = f.(string)
	}

	var searchOptions stats.SearchOptions
	if search, ok := config["search"].(map[string]interface{}); ok {
		if size, ok := search["size"].(int); ok {
			searchOptions.Size = size
		} else {
			searchOptions.Size = 1000
		}

		if runName, ok := search["run_name"].(string); ok {
			searchOptions.RunName = runName
		} else {
			searchOptions.RunName = "run"
		}
	}

	params := map[string]float64{"k": 10, "lambda": 0.5}
	if p, ok := config["params"].(map[string]float64); ok {
		params = p
	}

	return stats.NewTerrierStatisticsSource(stats.TerrierParameters(params), stats.TerrierField(field), stats.TerrierPropertiesPath(propsFile), stats.TerrierSearchOptions(searchOptions))
}
