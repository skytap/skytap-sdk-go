TEST?=$$(go list ./... | grep -v 'vendor')
GOFMT_FILES?=$$(find . -name '*.go')
GOIMPORT_FILES?=$$(find . -type f -name '*.go' -not -path './vendor/*')
PKG_NAME=skytap

default: build

build: fmtcheck
	go mod vendor
	go install ./skytap

test: fmtcheck
	go mod vendor
	go test -i $(TEST) || exit 1
	echo $(TEST) | \
		xargs -t -n4 go test $(TESTARGS) -timeout=30s -parallel=4 -v

vet:
	@echo "go vet ./skytap"
	@go vet $$(go list ./...) ; if [ $$? -eq 1 ]; then \
		echo ""; \
		echo "Vet found suspicious constructs. Please check the reported constructs"; \
		echo "and fix them if necessary before submitting the code for review."; \
		exit 1; \
	fi

fmt:
	gofmt -w $(GOFMT_FILES)

fmtcheck:
	@sh -c "'$(CURDIR)/scripts/gofmtcheck.sh'"

test-compile:
	@if [ "$(TEST)" = "./..." ]; then \
		echo "ERROR: Set TEST to a specific package. For example,"; \
		echo "  make test-compile TEST=./$(PKG_NAME)"; \
		exit 1; \
	fi
	go test -c $(TEST) $(TESTARGS)

lint:
	golint skytap

imports:
	goimports -w $(GOIMPORT_FILES)

.PHONY: build test vet fmt fmtcheck test-compile lint imports

