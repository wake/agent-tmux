.PHONY: build test lint clean

BINARY=tsm

build:
	go build -o $(BINARY) ./cmd/tsm

test:
	go test ./... -v -race

test-cover:
	go test ./... -coverprofile=coverage.out -race
	go tool cover -html=coverage.out

lint:
	go vet ./...

clean:
	rm -f $(BINARY) coverage.out

run: build
	./$(BINARY)
