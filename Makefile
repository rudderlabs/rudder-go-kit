.PHONY: help default test test-run test-teardown generate lint fmt

GO=go
LDFLAGS?=-s -w
TESTFILE=_testok

# go tools versions
GOLANGCI=github.com/golangci/golangci-lint/cmd/golangci-lint@v1.57.1
gofumpt=mvdan.cc/gofumpt@latest
govulncheck=golang.org/x/vuln/cmd/govulncheck@latest
goimports=golang.org/x/tools/cmd/goimports@latest
mockgen=go.uber.org/mock/mockgen@v0.4.0
gotestsum=gotest.tools/gotestsum@v1.11.0
protoc-gen-go=google.golang.org/protobuf/cmd/protoc-gen-go@v1.33.0
protoc-gen-go-grpc=google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.3.0

default: lint

generate: install-tools
	$(GO) generate ./...

test: install-tools test-run test-teardown

test-run: ## Run all unit tests
ifeq ($(filter 1,$(debug) $(RUNNER_DEBUG)),)
	$(eval TEST_CMD = SLOW=0 gotestsum --format pkgname-and-test-fails --)
	$(eval TEST_OPTIONS = -race -p=1 -v -failfast -shuffle=on -coverprofile=profile.out -covermode=atomic -coverpkg=./... -vet=all --timeout=15m)
else
	$(eval TEST_CMD = SLOW=0 go test)
	$(eval TEST_OPTIONS = -race -p=1 -v -failfast -shuffle=on -coverprofile=profile.out -covermode=atomic -coverpkg=./... -vet=all --timeout=15m)
endif
ifdef package
ifdef exclude
	$(eval FILES = `go list ./$(package)/... | egrep -iv '$(exclude)'`)
	$(TEST_CMD) -count=1 $(TEST_OPTIONS) $(FILES) && touch $(TESTFILE) || true
else
	$(TEST_CMD) $(TEST_OPTIONS) ./$(package)/... && touch $(TESTFILE) || true
endif
else ifdef exclude
	$(eval FILES = `go list ./... | egrep -iv '$(exclude)'`)
	$(TEST_CMD) -count=1 $(TEST_OPTIONS) $(FILES) && touch $(TESTFILE) || true
else
	$(TEST_CMD) -count=1 $(TEST_OPTIONS) ./... && touch $(TESTFILE) || true
endif

test-teardown:
	@if [ -f "$(TESTFILE)" ]; then \
    	echo "Tests passed, tearing down..." ;\
		rm -f $(TESTFILE) ;\
		echo "mode: atomic" > coverage.txt ;\
		find . -name "profile.out" | while read file; do grep -v 'mode: atomic' $${file} >> coverage.txt; rm -f $${file}; done ;\
	else \
    	rm -f coverage.txt coverage.html ; find . -name "profile.out" | xargs rm -f ;\
		echo "Tests failed :-(" ;\
		exit 1 ;\
	fi

coverage:
	go tool cover -html=coverage.txt -o coverage.html

test-with-coverage: test coverage

help: ## Show the available commands
	@grep -E '^[0-9a-zA-Z_-]+:.*?## .*$$' ./Makefile | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

install-tools:
	$(GO) install $(gotestsum)
	$(GO) install $(mockgen)
	$(GO) install $(protoc-gen-go)
	$(GO) install $(protoc-gen-go-grpc)

.PHONY: lint
lint: fmt ## Run linters on all go files
	$(GO) run $(GOLANGCI) run -v

.PHONY: fmt
fmt: install-tools ## Formats all go files
	$(GO) run $(govulncheck) ./...
	$(GO) run $(gofumpt) -l -w -extra  .
	find . -type f -name '*.go' -exec grep -L -E 'Code generated by .*\. DO NOT EDIT.' {} + | xargs $(GO) run $(goimports) -format-only -w -local=github.com/rudderlabs
