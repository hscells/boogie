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

The most important abstraction is the statistic source. A boogie pipeline is not concerned with how documents are
stored or the structure of the index; only how to retrieve documents. In this way, boogie separates
how you choose to store your documents from how you get your experiments done.

## Installation

boogie can be installed with `go install`.

```bash
go install github.com/hscells/boogie/cmd/boogie
```

## Usage

For command line help, see `boogie --help`.

command line usage:

```bash
boogie --pipeline pipeline.json
```

 - `--pipeline`; the path to a boogie pipeline file which will be used to construct a groove pipeline.
 - `--logfile` (optional); the path to a logfile to output logs to.

**Important:** Queries require a specific format that is used by groove. Each query file must contain one query, and the
name of the file must be the topic for that query. For example, if topic 1 contains the query:

```
green eggs and ham
```

Then the file must be named `1`. Groove uses this to process evaluation and result files.

## DSL

boogie uses a domain specific language (DSL) for creating [groove](https://github.com/hscells/groove) pipelines.
The boogie DSL looks like a regular JSON file (although JSON is notoriously bad for these types of things, I think in this case it is OK). The example below provides a pipeline for running a simple IR experiment - run some queries in a search engine and evaluate them:

```json
{
  "query": {
    "format": "medline",
    "path": "path/to/queries"
  },
  "statistic": {
    "source": "elasticsearch"
    ...
  },
  "evaluations": [
    "precision",
    "recall",
    "f1"
  ]
  "output": {
    "evaluations": {
      "qrels": "medline.qrels",
      "formats": [
        {
          "format": "json",
          "filename": "medline_bool.json"
        }
      ]
    },
    "trec_results": {
      "output": "medline_bool.results"
    }
  }
}
```

There are currently 11 different top-level configuration items that may or may not integrate with each other. I have tried my best to describe each of these items and how they can interact with each other.

### Query (`query`)

Query formats are specified using the `format`, the different query formats and options are detailed below. The path to
your queries should be specified using `path`.

#### `medline`

 - `mapping`: Specify a field mapping in the same format as when loading a field mapping into
 [transmute](https://github.com/hscells/transmute).

#### `pubmed`

The options for the `pubmed` format are the same as `medline`.

#### `keyword`

A keyword query (just one string of characters per file). No additional options may be specified.

### Statistic (`statistic`)

Statistic sources provide common information retrieval methods. They are specified using `source`. The source component
and options are detailed below. groove/boogie does not attempt to configure information retrieval systems (e.g. sources
for statistics), only attempt to wrap them in some way. For this reason, you should read how to set up these systems
before using boogie. A statistic source can only be configured if `query` has been configured. 

There are currently three configured statistic sources: Elasticsearch, Terrier, and Entrez.

#### `elasticsearch`

 - `hosts`: Specify a list of Elasticsearch urls (e.g. http://example.com:9200)
 - `index`: Elasticsearch index to run experiments on.
 - `document_type`: Elasticsearch document type.
 - `field`: Field to search on (for keyword queries).
 - `analyser`: Specify a preconfigured analyser for term vectors/analyse transformation.
 - `analyse_field`: Specify the field to be analysed for term vectors/analyse transformation.
 - `scroll`: Specify whether to scroll or not (true/false).

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

#### `entrez`

 - `email`: Email of the account using Entrez.
 - `tool`: Tool name accessing Entrez.
 - `key`: (optional) Key parameter of Entrez (to increase rate limit).

#### Universal options:

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

### Query Rewrites (`rewrite`)

Rewrites are a different type of transformation in that they can be applied in multiple ways to a query. These are useful for creating query variations or exploring the space of possible queries. Currently the only use for this is in the query chain machine learning model. But the variations could, for example, just be output to a directory (I'm too lazy to do this).

The possible rewrites that are available are:

 - `logical_operator_replacement`: Replace ORs with ANDs and ANDs with ORs
 - `adj_range`: Modify the distance of adjacency operators.
 - `adj_replacement`: Replace adjacency operators with AND operators.
 - `mesh_explosion`: Explode/Unexplode a MeSH keyword.
 - `mesh_parent`: Move a MeSH keyword up one level in the ontology.
 - `field_restrictions`: Permute the fields being searched on.

### Measurements (`measurements`)

Measurements are methods that apply a calculation to a query using a statistics source. All measurements return a
floating point number. This component accepts a list of measurements:

 - `term_count` - Total number of query terms.
 - `avg_ictf` - Average inverse collection term frequency.
 - `avg_idf` - Average inverse document frequency.
 - `sum_idf` - Sum inverse document frequency.
 - `max_idf` - Max inverse document frequency.
 - `std_idf` - Standard Deviation inverse document frequency.
 - `sum_cqs` - Sum Collection Query Similarity
 - `max_cqs` - Max Collection Query Similarity.
 - `avg_cqs` - Average Collection Query Similarity.
 - `scs` - Simplified Clarity Score.
 - `query_scope` - Query Scope.
 - `wig` - Weighted Information Gain.
 - `weg` - Weighted Entropy Gain.
 - `ncq` - Normalised Query Commitment.
 - `clarity_score` - Clarity Score.
 - `retrieval_size` - Total number of documents retrieved.
 - `boolean_clauses` - Number of clauses in Boolean query.
 - `boolean_keywords` - Number of keywords in Boolean query.
 - `boolean_fields` - Number of fields in Boolean query.
 - `boolean_truncated` - Number of wildcard keywords in Boolean query.
 - `mesh_keywords` - Number of MeSH keywords in Boolean query.
 - `mesh_exploded` - Number of Exploded MeSH keywords in Boolean query.
 - `mesh_non_exploded` - Number of Non-Exploded MeSH keywords in Boolean Query
 -  `mesh_avg_depth` - Average depth of MeSH keywords in ontology in Boolean query.
 -  `mesh_max_depth` - Maximum depth of MeSH keywords in ontology in Boolean query.

Measurements can just be output to a file, or be used as inputs to machine learning (for example feature engineering; see below).

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
 - `f05_measure`: F-beta 0.5
 - `f1_measure`: F-beta 1
 - `f3_measure`: F-beta 3
 - `wss`: Work Saved over Sampling

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

### Machine Learning (`learning`)

Machine learning is kind of new in boogie and it's still not perfect, but at the moment there is some learning to rank being implemented. 

 - `model`: Which machine learning model to use (currently available: `query_chain`).
 - `options`: Additional model-specific options (see below).
 - `train`: Options for training.
 - `test`: Options for testing.
 - `generate`: Options for generating data.

Even if a model does not have configuration options for training, testing, or generating, it tells the pipeline which operation(s) to perform.

#### `query_chain`

##### Options:

 - `candidate_selector`: one of `ltr_svmrank`, `ltr_quickrank`, or `reinforcement` (only `ltr_quickrank` is fully implemented).
    - `ltr_svmrank` requires `model_file` to be configured here.
    - `ltr_quickrank` requires `binary` to be configured here, as well as any arguments to quickrank (see: https://github.com/hpclab/quickrank)
 
##### Generate:

Query chain generate requires that a query, statistic source, measurements, rewrite, evaluation, and output (qrels) is configured.

 - `output`: Path to generate features to.

## Extending

Adding a query format, statistics source, preprocessing step, measurement, or output format requires firstly to
implement the corresponding [groove](https://github.com/hscells/groove) interface. Once an interface has been
implemented, it can be added to boogie by registering it in the [config](config.go).

I am open to contributions, but having said that I would not be contributing at this point in time unless it was to a really stable API like evaluation or measurements.

## Citing

If you use this work for scientific publication, please reference

```
@inproceedings{scells2018framework,
 author = {Scells, Harrisen and Locke, Daniel and Zuccon, Guido},
 title = {An Information Retrieval Experiment Framework for Domain Specific Applications},
 booktitle = {The 41st International ACM SIGIR Conference on Research \&\#38; Development in Information Retrieval},
 series = {SIGIR '18},
 year = {2018},
} 
```

## Logo

The Go gopher was created by [Renee French](https://reneefrench.blogspot.com/), licensed under
[Creative Commons 3.0 Attributions license](https://creativecommons.org/licenses/by/3.0/).
