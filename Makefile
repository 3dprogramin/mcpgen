.PHONY: build install test vet fmt fmt-check check demo snapshot clean

build:
	go build -o mcpgen .

install:
	go install .

test:
	go test -race ./...

vet:
	go vet ./...

fmt:
	gofmt -w .

fmt-check:
	@unformatted=$$(gofmt -l .); \
	if [ -n "$$unformatted" ]; then echo "Not gofmt'd:"; echo "$$unformatted"; exit 1; fi

# Run everything CI runs.
check: fmt-check vet test build

# Render the README demo GIF (requires https://github.com/charmbracelet/vhs).
demo: build
	vhs demo.tape

# Build a local release snapshot without publishing (requires goreleaser).
snapshot:
	goreleaser release --snapshot --clean

clean:
	rm -f mcpgen
	rm -rf dist/
