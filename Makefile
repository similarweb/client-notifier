# Go parameters
GOCMD=go
GOTEST=$(GOCMD) test
GOTESTRACE=$(GOTEST) -race
GOFMT=$(GOCMD)fmt


test: ## Run tests for the project
		$(GOTEST) -count=1 -coverprofile=cover.out -short -cover -failfast ./...

test-race: ## Run tests for the project (while detecting race conditions)
		$(GOTESTRACE) -coverprofile=cover.out -short -cover -failfast ./...

test-html: test ## Run tests with HTML for the project
		$(GOTOOL) cover -html=cover.out
	
gofmt: ## gofmt code formating
	@echo Running go formating with the following command:
	$(GOFMT) -e -s -w .

fmt-validator: ## Validate go format
	@echo checking gofmt...
	@res=$$($(GOFMT) -d -e -s $$(find . -type d \( -path ./src/vendor \) -prune -o -name '*.go' -print)); \
	if [ -n "$${res}" ]; then \
		echo checking gofmt fail... ; \
		echo "$${res}"; \
		exit 1; \
	else \
		echo Your code formating is according gofmt standards; \
	fi

checks-validator: fmt-validator ## Run all validations

help: ## Show Help menu
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
