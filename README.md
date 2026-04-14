# large_rdf_bench_result_json_format

Converts [Large RDF Bench](https://github.com/dice-group/LargeRDFBench) query result files into [W3C SPARQL JSON Result Format](https://www.w3.org/TR/sparql11-results-json/).

## Dependencies
- [Go version 1.25.7](https://go.dev/)
- [GNU Make](https://www.gnu.org/software/make/)

## Usage

Build the binary:

```zsh
make build
```

The binary will be built in `./bin/large_rdf_bench_result_json_format`.

Run it:

```
Usage of ./bin/large_rdf_bench_result_json_format:
  -folder string
    	input directory (required)
  -o string
    	output directory (required)
```
