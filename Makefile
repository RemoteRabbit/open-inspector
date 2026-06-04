.PHONY: all build test lint fmt tidy clean run license license-check license-fix pre-commit-install pre-commit

BIN            := bin/open-inspector
PKG            := ./...
GOFLAGS        ?=
ADDLICENSE     := go run github.com/google/addlicense@v1.2.0
LICENSE_HEADER := .licenseheader.tmpl
# Directories containing source files that must carry the MPL header.
# Extend this list as new top-level source dirs are added.
LICENSE_PATHS  := cmd pkg

all: fmt lint license-fix test build

build:
	@mkdir -p bin
	go build $(GOFLAGS) -o $(BIN) ./cmd/open-inspector

test:
	go test $(GOFLAGS) -race -count=1 $(PKG)

test-update:
	go test ./pkg/config -run TestLoad_Snapshots -update

lint:
	golangci-lint run

fmt:
	gofmt -s -w $(LICENSE_PATHS)
	go vet $(PKG)

tidy:
	go mod tidy

run: build
	$(BIN)

clean:
	rm -rf bin

# Add the MPL-2.0 header to any source file that is missing it.
license:
	$(ADDLICENSE) -f $(LICENSE_HEADER) $(LICENSE_PATHS)

# Fail (non-zero exit) if any source file is missing the MPL-2.0 header.
# Used by CI.
license-check:
	$(ADDLICENSE) -check -f $(LICENSE_HEADER) $(LICENSE_PATHS)

# Run license-check; if it fails, auto-apply headers and re-check so a
# locally-missing header doesn't abort `make all`.
license-fix:
	@$(ADDLICENSE) -check -f $(LICENSE_HEADER) $(LICENSE_PATHS) || { \
		echo "license-check failed; applying headers and retrying..."; \
		$(ADDLICENSE) -f $(LICENSE_HEADER) $(LICENSE_PATHS) && \
		$(ADDLICENSE) -check -f $(LICENSE_HEADER) $(LICENSE_PATHS); \
	}

# Install pre-commit + pre-push hooks defined in .pre-commit-config.yaml.
pre-commit-install:
	pre-commit install --install-hooks
	pre-commit install --hook-type pre-push

# Run every hook against every tracked file (useful in CI and after big
# edits).
pre-commit:
	pre-commit run --all-files
