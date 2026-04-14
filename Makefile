BINARY := bin/large_rdf_bench_cleanup
COVERAGE := 100

.PHONY:  lint build install

build: install
	go build -o $(BINARY) main.go

install:
	go mod download

lint:
	go tool golangci-lint run
	go tool golangci-lint fmt

ci: lint test

clean:
	rm -f $(BINARY)
