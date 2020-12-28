test:
	@./go.test.sh
.PHONY: test

coverage:
	@./go.coverage.sh
.PHONY: coverage

generate:
	go generate ./...
.PHONY: generate

download_schema:
	go run ./cmd/dltl -f api.tl -o _schema/telegram.tl
.PHONY: download_schema-schema

check_generated: generate
	git diff --exit-code
.PHONY: check_generated

fuzz_telegram:
	go run github.com/dvyukov/go-fuzz/go-fuzz -bin telegram/telegram-fuzz.zip -workdir _fuzz/handle_message
.PHONY: fuzz_telegram

fuzz_telegram_build:
	cd telegram && go run github.com/dvyukov/go-fuzz/go-fuzz-build -func FuzzHandleMessage -tags fuzz -o telegram-fuzz.zip
.PHONY: fuzz_telegram_build

fuzz_telegram_clear:
	rm -f _fuzz/handle_message/crashers/*
	rm -f _fuzz/handle_message/suppressions/*
.PHONY: fuzz_telegram_clear
