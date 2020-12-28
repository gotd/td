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

fuzz_rsa:
	go run github.com/dvyukov/go-fuzz/go-fuzz -bin internal/crypto/rsa-fuzz.zip -workdir _fuzz/rsa
.PHONY: fuzz_rsa

fuzz_rsa_build:
	cd internal/crypto && go run github.com/dvyukov/go-fuzz/go-fuzz-build -func FuzzRSA -tags fuzz -o rsa-fuzz.zip
.PHONY: fuzz_rsa_build

fuzz_rsa_clear:
	rm -f _fuzz/rsa/crashers/*
	rm -f _fuzz/rsa/suppressions/*
.PHONY: fuzz_rsa_clear
