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

boogie uses a domain specific language (DSL) for creating [groove](https://github.com/hscells/boogie) pipelines.
The boogie dsl looks like a regular JSON file:

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
    "boolean_query_count"
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
formats. These components are reflected in the top-level keys in the dsl. To see what sources can be used, see
[config.go](config.go) where the sources are registered. To add custom measurements or sources, they must be registered
in this file.