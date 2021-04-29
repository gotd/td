test:
	@./go.test.sh
.PHONY: test

coverage:
	@./go.coverage.sh
.PHONY: coverage

generate:
	go generate
	go generate ./...
.PHONY: generate

download_schema:
	go run ./cmd/dltl -f api.tl -o _schema/telegram.tl -merge _schema/help_config_simple.tl
.PHONY: download_schema

download_public_keys:
	go run ./cmd/dlkey -o internal/mtproto/_data/public_keys.pem
.PHONY: download_public_keys

download_e2e_schema:
	go run ./cmd/dltl -base https://raw.githubusercontent.com/tdlib/td -branch master -dir td/generate/scheme -f secret_api.tl -o _schema/encrypted.tl
.PHONY: download_e2e_schema

check_generated: generate
	git diff --exit-code
.PHONY: check_generated

fuzz_telegram:
	cd internal/mtproto && go test -test.fuzztime 10m -test.fuzzcachedir ../../_fuzz --fuzz .
.PHONY: fuzz_telegram

fuzz_rsa:
	cd internal/crypto && go test -test.fuzztime 10m -test.fuzzcachedir ../../_fuzz --fuzz .
.PHONY: fuzz_rsa

fuzz_flow:
	cd internal/exchange && go test -test.fuzztime 10m -test.fuzzcachedir ../../_fuzz --fuzz .
.PHONY: fuzz_flow
