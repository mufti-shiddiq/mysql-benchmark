.PHONY: build test lint run clean

build:
	go build -o mysql-benchmark ./cmd/mysql-benchmark

test:
	go test ./...

lint:
	go vet ./...

run:
	go run ./cmd/mysql-benchmark

clean:
	rm -f mysql-benchmark
