package boogie

import (
	"bytes"
	"encoding/gob"
	"github.com/hscells/cui2vec"
	"github.com/hscells/groove/analysis"
	"github.com/hscells/groove/combinator"
	"github.com/hscells/groove/eval"
	"github.com/hscells/groove/learning"
	"github.com/hscells/groove/output"
	"github.com/hscells/groove/preprocess"
	"github.com/hscells/groove/query"
	"github.com/hscells/groove/stats"
	"github.com/hscells/quickumlsrest"
	"github.com/hscells/transmute/pipeline"
	"github.com/hscells/trecresults"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

var (
	querySourceMapping                 = map[string]query.QueriesSource{}
	statisticSourceMapping             = map[string]stats.StatisticsSource{}
	preprocessorMapping                = map[string]preprocess.QueryProcessor{}
	transformationMappingBoolean       = map[string]preprocess.BooleanTransformation{}
	transformationMappingElasticsearch = map[string]preprocess.ElasticsearchTransformation{}
	measurementMapping                 = map[string]analysis.Measurement{}
	measurementFormatters              = map[string]output.MeasurementFormatter{}
	evaluationMapping                  = map[string]eval.Evaluator{}
	evaluationFormatters               = map[string]output.EvaluationFormatter{}
	rewriteTransformationMapping       = map[string]learning.Transformation{}
	modelMapping                       = map[string]learning.Model{}
)

// RegisterQuerySource registers a query source.
func RegisterQuerySource(name string, source query.QueriesSource) {
	querySourceMapping[name] = source
}

// RegisterStatisticSource registers a statistic source.
func RegisterStatisticSource(name string, source stats.StatisticsSource) {
	statisticSourceMapping[name] = source
}

// RegisterPreprocessor registers a preprocessor.
func RegisterPreprocessor(name string, preprocess preprocess.QueryProcessor) {
	preprocessorMapping[name] = preprocess
}

// RegisterTransformationBoolean registers a Boolean query transformation.
func RegisterTransformationBoolean(name string, transformation preprocess.BooleanTransformation) {
	transformationMappingBoolean[name] = transformation
}

// RegisterTransformationElasticsearch registers an Elasticsearch transformation.
func RegisterTransformationElasticsearch(name string, transformation preprocess.ElasticsearchTransformation) {
	transformationMappingElasticsearch[name] = transformation
}

// RegisterMeasurement registers a measurement.
func RegisterMeasurement(name string, measurement analysis.Measurement) {
	measurementMapping[name] = measurement
}

// RegisterMeasurementFormatter registers an output formatter.
func RegisterMeasurementFormatter(name string, formatter output.MeasurementFormatter) {
	measurementFormatters[name] = formatter
}

// RegisterEvaluator registers a measurement.
func RegisterEvaluator(name string, evaluator eval.Evaluator) {
	evaluationMapping[name] = evaluator
}

// RegisterEvaluationFormatter registers an output formatter.
func RegisterEvaluationFormatter(name string, formatter output.EvaluationFormatter) {
	evaluationFormatters[name] = formatter
}

// RegisterRewriteTransformation registers a rewrite transformation.
func RegisterRewriteTransformation(name string, transformation learning.Transformation) {
	rewriteTransformationMapping[name] = transformation
}

func RegisterCui2VecTransformation(dsl Pipeline) error {
	if len(dsl.Utilities.CUI2vec) > 0 && len(dsl.Utilities.CUIMapping) > 0 && len(dsl.Utilities.QuickUMLSCache) > 0 {
		var (
			embeddings cui2vec.Embeddings
			err        error
		)

		f, err := os.OpenFile(dsl.Utilities.CUI2vec, os.O_RDONLY, 0644)
		if err != nil {
			return err
		}

		// First, load the embeddings file.
		// If the file is a csv, then we can try loading it as such.
		if strings.Contains(dsl.Utilities.CUI2vec, ".csv") {
			embeddings, err = cui2vec.NewUncompressedEmbeddings(f, dsl.Utilities.CUI2vecSkip, ',')
			if err != nil {
				return err
			}
		} else { // Otherwise, we assume the file is a binary file.
			embeddings, err = cui2vec.NewPrecomputedEmbeddings(f)
			if err != nil {
				return err
			}
		}

		// Secondly, load the file that will perform the mapping from CUI->string.
		mapping, err := cui2vec.LoadCUIMapping(dsl.Utilities.CUIMapping)
		if err != nil {
			return err
		}

		// Lastly, load the file that contains cached cui similarity mappings.
		f, err = os.OpenFile(dsl.Utilities.QuickUMLSCache, os.O_RDONLY, os.ModePerm)
		if err != nil {
			return err
		}
		var cache quickumlsrest.Cache
		err = gob.NewDecoder(f).Decode(&cache)
		if err != nil {
			return err
		}

		// Finally, register a client that will communicate to the QuickUMLS REST API.
		//quickumls := quickumlsrest.NewClient(dsl.Utilities.QuickUMLSRest)
		RegisterRewriteTransformation("cui2vec_expansion", learning.Newcui2vecExpansionTransformer(embeddings, mapping, cache))
	}
	return nil
}

// RegisterQueryChainCandidateSelector registers a query chain candidate selector.
func RegisterModel(name string, model learning.Model) {
	modelMapping[name] = model
}

// NewOracleQueryChainCandidateSelector creates a new oracle query chain candidate selector.
func NewOracleQueryChainCandidateSelector(source string, qrels string) learning.OracleQueryChainCandidateSelector {
	b, err := ioutil.ReadFile(qrels)
	if err != nil {
		panic(err)
	}
	q, err := trecresults.QrelsFromReader(bytes.NewReader(b))
	if err != nil {
		panic(err)
	}

	if ss, ok := statisticSourceMapping[source]; ok {
		// TODO the cache should be able to be configured.
		return learning.NewOracleQueryChainCandidateSelector(ss, q, combinator.NewMapQueryCache())
	}

	log.Fatal("could not create oracle query chain candidate selector")
	return learning.OracleQueryChainCandidateSelector{}
}

// NewKeywordQuerySource creates a "keyword query" query source.
func NewKeywordQuerySource(options map[string]interface{}) query.QueriesSource {
	fields := []string{"text"}
	if optionFields, ok := options["fields"].([]interface{}); ok {
		fields = make([]string, len(optionFields))
		for i, f := range optionFields {
			fields[i] = f.(string)
		}
	}

	return query.NewKeywordQuerySource(fields...)
}

// NewTransmuteQuerySource creates a new transmute query source for PubMed/Medline queries.
func NewTransmuteQuerySource(p pipeline.TransmutePipeline, options map[string]interface{}) query.QueriesSource {
	if _, ok := options["mapping"]; ok {
		if mapping, ok := options["mapping"].(map[string][]string); ok {
			p.Options.FieldMapping = mapping
		}
	}

	return query.NewTransmuteQuerySource(p)
}

// NewElasticsearchStatisticsSource attempts to create an Elasticsearch statistics source from a configuration mapping.
// It also tries to set some defaults for fields in case some are not specified, but they will not be sensible.
func NewElasticsearchStatisticsSource(config map[string]interface{}) (*stats.ElasticsearchStatisticsSource, error) {
	var esHosts []string
	documentType := "doc"
	index := "index"
	scroll := false

	if hosts, ok := config["hosts"]; ok {
		for _, host := range hosts.([]interface{}) {
			esHosts = append(esHosts, host.(string))
		}
	} else {
		esHosts = []string{"http://localhost:9200"}
	}

	var searchOptions stats.SearchOptions
	if search, ok := config["search"].(map[string]interface{}); ok {
		if size, ok := search["size"].(float64); ok {
			searchOptions.Size = int(size)
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

	if d, ok := config["document_type"]; ok {
		documentType = d.(string)
	}

	if i, ok := config["index"]; ok {
		index = i.(string)
	}

	if s, ok := config["scroll"]; ok {
		scroll = s.(bool)
	}

	return stats.NewElasticsearchStatisticsSource(
		stats.ElasticsearchDocumentType(documentType),
		stats.ElasticsearchIndex(index),
		stats.ElasticsearchHosts(esHosts...),
		stats.ElasticsearchParameters(params),
		stats.ElasticsearchSearchOptions(searchOptions),
		stats.ElasticsearchScroll(scroll))
}

func NewEntrezStatisticsSource(config map[string]interface{}, options ...func(source *stats.EntrezStatisticsSource)) (stats.EntrezStatisticsSource, error) {
	var (
		tool, email, key string
		rank             = false
	)

	if d, ok := config["tool"]; ok {
		tool = d.(string)
	}

	if d, ok := config["email"]; ok {
		email = d.(string)
	}

	if d, ok := config["key"]; ok {
		key = d.(string)
	}

	if d, ok := config["rank"]; ok {
		rank = d.(bool)
	}

	var searchOptions stats.SearchOptions
	if search, ok := config["search"].(map[string]interface{}); ok {
		if size, ok := search["size"].(float64); ok {
			searchOptions.Size = int(size)
		} else {
			searchOptions.Size = 1000
		}

		if runName, ok := search["run_name"].(string); ok {
			searchOptions.RunName = runName
		} else {
			searchOptions.RunName = "run"
		}
	}

	e, err := stats.NewEntrezStatisticsSource(
		stats.EntrezAPIKey(key),
		stats.EntrezEmail(email),
		stats.EntrezTool(tool),
		stats.EntrezOptions(searchOptions),
		stats.EntrezRank(rank))
	if err != nil {
		return e, nil
	}

	for _, option := range options {
		option(&e)
	}

	return e, nil
}
