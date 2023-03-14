.PHONY: help default test test-run test-teardown generate lint fmt

GO=go
LDFLAGS?=-s -w
TESTFILE=_testok

default: lint

generate: install-tools
	$(GO) generate ./...

test: install-tools test-run test-teardown

test-run: ## Run all unit tests
ifeq ($(filter 1,$(debug) $(RUNNER_DEBUG)),)
	$(eval TEST_CMD = SLOW=0 gotestsum --format pkgname-and-test-fails --)
	$(eval TEST_OPTIONS = -p=1 -v -failfast -shuffle=on -coverprofile=profile.out -covermode=count -coverpkg=./... -vet=all --timeout=15m)
else
	$(eval TEST_CMD = SLOW=0 go test)
	$(eval TEST_OPTIONS = -p=1 -v -failfast -shuffle=on -coverprofile=profile.out -covermode=count -coverpkg=./... -vet=all --timeout=15m)
endif
ifdef package
	$(TEST_CMD) $(TEST_OPTIONS) $(package) && touch $(TESTFILE) || true
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
	go install github.com/golang/mock/mockgen@v1.6.0
	go install mvdan.cc/gofumpt@latest
	go install gotest.tools/gotestsum@v1.8.2

.PHONY: lint
lint: fmt ## Run linters on all go files
	docker run --rm -v $(shell pwd):/app:ro -w /app golangci/golangci-lint:v1.51.1 bash -e -c \
		'golangci-lint run -v --timeout 5m'

.PHONY: fmt
fmt: install-tools ## Formats all go files
	gofumpt -l -w -extra  .
