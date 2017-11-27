<img height="200px" src="gopher.png" alt="gopher" align="right"/>

# boogie

[![GoDoc](https://godoc.org/github.com/hscells/boogie?status.svg)](https://godoc.org/github.com/hscells/boogie)
[![Go Report Card](https://goreportcard.com/badge/github.com/hscells/boogie)](https://goreportcard.com/report/github.com/hscells/boogie)

_DSL front-end for [groove](https://github.com/hscells/groove)_

## Installation

boogie can be installed with `go get`.

```bash
go get -u github.com/hscells/boogie
```

## Usage

For command line help, see `boogie --help`.

command line usage:

```bash
boogie --queries ./medline --pipeline pipeline.json
```

 - `--queries` is the path to a directory of queries that will be analysed by groove.
 - `--pipeline` is the path to a boogie pipeline file which will be used to construct a groove pipeline.

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

### Query Transformations (`transformations`)

Query transformations are operations that change queries beyond simple string manipulation. For instance, a
transformation can simplify a query, or replace Boolean operators. The output directory and a list of transformations
can be specified. If the directory is not present, no queries will be output.

 - `output`: Directory to output transformed queries to.
 - `operations`: List of transformations to apply (see below).

The possible query transformation operations are listed as follows:

 - `simplify`: Simplify a Boolean query to just "and" and "or" operators.

Operations are applied in the order specified.

### Measurements (`measurements`)

Measurements are methods that apply a calculation to a query using a statistics source. All measurements return a
floating point number. This component accepts a list of preprocessors:

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

### Output (`output`)

An output specifies how experiments are to be formatted and what file to write them to. The `output` component comprises
a list of outputs. Each output contains a `format` field and a `filename` field. The `filename` field tells the pipeline
where to write the file, and the `format` is the format of the file. The formats are described below.

 - `json`: JSON formatting.
 - `csv`: Comma separated formatting.

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