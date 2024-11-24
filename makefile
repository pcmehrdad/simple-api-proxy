.PHONY: run build clean

run:
	go run ./cmd/processor/main.go

build:
	go build -o bin/api-proxy ./cmd/processor

clean:
	rm -f bin/api-proxy