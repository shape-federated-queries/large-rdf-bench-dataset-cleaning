FROM docker.io/golang:1.25 AS builder

WORKDIR /src
COPY go.mod go.sum Makefile main.go qleverlfile_template ./
RUN make build

FROM docker.io/adfreiburg/qlever:latest

WORKDIR /workspace
COPY --from=builder /src/bin/large_rdf_bench_cleanup /usr/local/bin/large_rdf_bench_cleanup

VOLUME ["/out"]
ENTRYPOINT ["/usr/local/bin/large_rdf_bench_cleanup", "-o", "/out"]
