# dataset-cleaning

Cleans [Large RDF Bench](https://github.com/dice-group/LargeRDFBench) RDF datasets with QLever.

## Dependencies
- [Go version 1.25.7](https://go.dev/)
- [GNU Make](https://www.gnu.org/software/make/)
- [Podman](https://podman.io/) (for container usage)

## Local usage

Build the binary:

```zsh
make build
```

The binary will be built in `./bin/large_rdf_bench_cleanup`.

Run it:

```
Usage of ./bin/large_rdf_bench_cleanup:
  -g string
    	input glob (required)
  -o string
    	output directory (required)
```

## Podman usage

Build image:

```zsh
podman build -t large-rdf-bench-dataset-cleaning:latest .
```

Run cleanup with output persisted in the `/out` volume path inside the container.  
`-o /out` is enforced by the image entrypoint, and you provide only `-g` at runtime.
The input path must be bind-mounted from the host:

```zsh
podman run --rm \
    --userns=keep-id \
    --user $(id -u):$(id -g) \
    -v ./out:/out:z \
    -v {path to the benchmark files}:/host-input:z \
    large-rdf-bench-dataset-cleaning:latest \
    -g "/host-input/*.nt"
```
