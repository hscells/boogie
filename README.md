<img height="200px" src="gopher.png" alt="gopher" align="right"/>

# boogie

[![GoDoc](https://godoc.org/github.com/hscells/boogie?status.svg)](https://godoc.org/github.com/hscells/boogie)
[![Go Report Card](https://goreportcard.com/badge/github.com/hscells/boogie)](https://goreportcard.com/report/github.com/hscells/boogie)

_DSL front-end for [groove](https://github.com/hscells/groove)_

Often, we would like to abstract away the way we perform measurements (e.g. query performance prediction) or the way we
transform queries (e.g. query expansion/reduction); and we would like to do these things in a repeatable, reproducible
manner. This is where boogie comes in: an experiment is represented as a pipeline of operations that cover the format
of queries, the source for statistics, the operations and measurements for each query, and how the experiment is to be
output. boogie translates a simple DSL syntax into a [groove](https://github.com/hscells/groove) pipeline. Both groove
and boogie are designed to be easily extendable and offer sane, simple abstractions.

The most important abstraction is the statistic source. A boogie pipeline does not worry itself with how documents are
stored or the structure of your index; only how to access the source of documents. In this way, boogie separates
how you choose to store your documents from how you get your experiments done. boogie down.

## Installation

boogie can be installed with `go install`.

```bash
go install github.com/hscells/boogie
```

## Usage

For command line help, see `boogie --help`.

command line usage:

```bash
boogie --queries ./medline --pipeline pipeline.json
```

 - `--queries`;  the path to a directory of queries that will be analysed by groove.
 - `--pipeline`; the path to a boogie pipeline file which will be used to construct a groove pipeline.
 - `--logfile` (optional); the path to a logfile to output logs to.

## DSL

boogie uses a domain specific language (DSL) for creating [groove](https://github.com/hscells/groove) pipelines.
The boogie DSL looks like a regular JSON file:

```json
{
  "query": {
    "source": "medline"
  },
  "statistic": {
    "source": "elasticsearch"
  },
  "measurements": [
    "keyword_count",
    "term_count",
    "avg_idf"
    "avg_ictf"
  ],
  "preprocess": [
    "lowercase"
  ],
  "output": [
    {
      "format": "json",
      "filename": "analysis.json"
    },
    {
      "format": "csv",
      "filename": "analysis.csv"
    }
  ]
}
```

There are four components to a groove pipeline: the query source, the statistics source, measurements, and output
formats. These components are reflected in the top-level keys in the DSL. Each of the components are described below.

### Query (`query`)

Query formats are specified using the `format`, the different query formats and options are detailed below.

#### `medline`

 - `mapping`: Specify a field mapping in the same format as when loading a field mapping into
 [transmute](https://github.com/hscells/transmute).

#### `pubmed`

The options for the `pubmed` format are the same as `medline`.

### Statistic (`statistic`)

Statistic sources provide common information retrieval methods. They are specified using `source`. The source component
and options are detailed below. groove/boogie does not attempt to configure information retrieval systems (e.g. sources
for statistics), only attempt to wrap them in some way. For this reason, you should read how to set up these systems
before using boogie.

#### `elasticsearch`

 - `hosts`: Specify a list of Elasticsearch urls (e.g. http://example.com:9200)
 - `document_type`: Elasticsearch document type.
 - `index`: Elasticsearch index to run experiments on.
 - `analyser`: Specify a preconfigured analyser for term vectors/analyse transformation.
 - `analyse_field`: Specify the field to be analysed for term vectors/analyse transformation.

Note: The `analyser` and `analyse_field` are to be used in the cases where you may have stemmed documents and stemmed
queries and wish to get a term vector for a pre-stemmed term in a query. To do this, point `analyse_field` to the
analysed field name (i.e. "keyword"). When only `analyser` is set, this defaults to normal behaviour. In this case, the
request to the Elasticsearch term vectors API will look like this:

```json
{
    "doc": {
        "text": "disease"
    },
    "term_statistics": true,
    "field_statistics": false,
    "offsets": false,
    "positions": false,
    "payloads": false,
    "fields": ["text.keyword"],
    "per_field_analyzer": {
        "text.keyword": ""
    }
}
```

If `analyse_field` is not specified, the `fields` and `per_field_analyzer` keys will whatever `field` is set to in the
pipeline (in the case above would be "text").

#### `terrier`

 - `properties`: Location of the terrier properties file.

#### Universal options:

 - `field`: Document field for analysis.
 - `params`: Map of parameter name to float value (e.g. k, lambda).
 - `search`: Search properties; `size` (maximum number of results to retrieve), `run_name` (name of the run for trec)

### Query Preprocessing (`preprocess`)

Preprocessing is performed before analysing a query. This component accepts a list of preprocessors:

 - `alphanum`: Remove non-alphanumeric characters.
 - `lowercase`: Transform uppercase characters to lowercase.
 - `strip_numbers`: Remove numbers.

### Query Transformations (`transformations`)

Query transformations are operations that change queries beyond simple string manipulation. For instance, a
transformation can simplify a query, or replace Boolean operators. The output directory and a list of transformations
can be specified. If the directory is not present, no queries will be output.

 - `output`: Directory to output transformed queries to.
 - `operations`: List of transformations to apply (see below).

The possible query transformation operations are listed as follows:

 - `simplify`: Simplify a Boolean query to just "and" and "or" operators.
 - `analyse`: Use Elasticsearch to analyse the query strings in the query.

Additionally, the following transformation can be used in conjunction with the Elasticsearch statistics source:

 - `analyse`: Run the analyser specified in `statistic` on the query.

Operations are applied in the order specified.

### Measurements (`measurements`)

Measurements are methods that apply a calculation to a query using a statistics source. All measurements return a
floating point number. This component accepts a list of preprocessors:

 - `term_count` - Total number of query terms.
 - `keyword_count` - Total number of keywords used in a Boolean query.
 - `boolean_query_count` - Total number of clauses in a Boolean query.
 - `avg_ictf` - Average inverse collection term frequency.
 - `avg_idf` - Average inverse document frequency.
 - `sum_idf` - Sum inverse document frequency.
 - `max_idf` - Max inverse document frequency.
 - `std_idf` - Standard Deviation inverse document frequency.
 - `sum_cqs` - Sum Collection Query Similarity
 - `max_cqs` - Max Collection Query Similarity.
 - `scs` - Simplified Clarity Score.
 - `query_scope` - Query Scope.
 - `wig` - Weighted Information Gain.
 - `weg` - Weighted Entropy Gain.
 - `ncq` - Normalised Query Commitment.
 - `clarity_score` - Clarity Score.

### Evaluation (`evaluation`)

Queries can be evaluated through different measures. To evaluate queries in the pipeline, use the `evaluation` key. Each
evaluation measurement comprises:

 - `evaluate`: The measure to evaluate each topic with.

The list of measures are as follows:

 - `num_ret`: Total number of retrieved documents.
 - `num_rel`: Total number of relevant documents (from qrels).
 - `num_rel_ret`: Total number of relevant documents that were retrieved.
 - `precision`: Ratio of relevant retrieved documents to retrieved documents.
 - `recall`: Ratio of relevant retrieved documents to relevant documents.

### Output (`output`)

An output specifies how experiments are to be formatted and what file to write them to. The `output` component comprises
a list of outputs. Each output can either of type `measurements`, `trec_results`, or `evaluations`.

For `measurements`, each item contains a `format` field and a `filename` field. The `filename` field tells the pipeline
where to write the file, and the `format` is the format of the file. The formats are described below.

 - `json`: JSON formatting.
 - `csv`: Comma separated formatting.

For `trec_results`, a filename must be specified using `output`:

 - `output`: Where to write trec-style results file to.

For `evaluations`, both the `qrels` file must be specified, and a list of formats similar to `measurements`; i.e.
a list of filename and format pairs:

 - `qrels`: Path to a trec-style qrels file.
 - `formats`: `format`, `filename` pairs.

The format of `evaluations` is currently only `json`.

### Trec Results (`trec`)

Tell groove whether to output trec result files or not. If the `output` key is present, the results will be output to
file specified.

 - `output`: Where to output trec result file to.

## Extending

Adding a query format, statistics source, preprocessing step, measurement, or output format requires firstly to
implement the corresponding [groove](https://github.com/hscells/groove) interface. Once an interface has been
implemented, it can be added to boogie by registering it in the [config](config.go).

## Logo

The Go gopher was created by [Renee French](https://reneefrench.blogspot.com/), licensed under
[Creative Commons 3.0 Attributions license](https://creativecommons.org/licenses/by/3.0/).