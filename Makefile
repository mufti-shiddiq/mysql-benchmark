.PHONY: build test lint run clean
.PHONY: release-snapshot

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
	rm -rf dist

release-snapshot:
	mkdir -p dist
	GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o dist/mysql-benchmark ./cmd/mysql-benchmark
	tar -czf dist/mysql-benchmark-linux-amd64.tar.gz -C dist mysql-benchmark
	rm -f dist/mysql-benchmark
	GOOS=linux GOARCH=arm64 go build -trimpath -ldflags="-s -w" -o dist/mysql-benchmark ./cmd/mysql-benchmark
	tar -czf dist/mysql-benchmark-linux-arm64.tar.gz -C dist mysql-benchmark
	rm -f dist/mysql-benchmark
