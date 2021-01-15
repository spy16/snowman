VERSION="`git describe --abbrev=0 --tags`"
COMMIT="`git rev-list -1 HEAD`"

all: clean fmt test build

fmt:
	@echo "Re-formatting..."
	@goimports -w -l ./

clean:
	@echo "Removing builds..."
	@rm -rf ./bin
	@echo "Tidy up go.mod..."
	@go mod tidy -v

test:
	@echo "Running go test..."
	@go test -cover ./...

benchmark:
	@echo "Running benchmarks..."
	@go test -bench=Benchmark -run=none ./...


build:
	@echo "Building snowman..."
	@go build -o ./bin/snowman ./cmd/snowman/

install:
	@echo "Installing snowman..."
	@go install ./cmd/snowman
