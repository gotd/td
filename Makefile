test:
	@./go.test.sh
.PHONY: test

coverage:
	@./go.coverage.sh
.PHONY: coverage

generate:
	go generate ./...
.PHONY: generate


check_generated: generate
	git diff --exit-code
.PHONY: check_generated
