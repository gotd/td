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
	go run ./cmd/dltl -base https://raw.githubusercontent.com/tdlib/td -branch master -dir td/generate/scheme -f telegram_api.tl -o _schema/tdlib.tl
	go run ./cmd/dltl -base https://raw.githubusercontent.com/telegramdesktop/tdesktop -branch dev -dir Telegram/SourceFiles/mtproto/scheme -f api.tl -o _schema/tdesktop.tl
	go run ./cmd/dltl -base https://raw.githubusercontent.com/telegramdesktop/tdesktop -branch dev -dir Telegram/SourceFiles/mtproto/scheme -f api.tl -merge _schema/legacy.tl -o _schema/telegram.tl
.PHONY: download_schema

download_public_keys:
	go run ./cmd/dlkey -o internal/mtproto/_data/public_keys.pem
.PHONY: download_public_keys

download_e2e_schema:
	go run ./cmd/dltl -f secret_api.tl -o _schema/encrypted.tl
.PHONY: download_e2e_schema

download_tdlib_schema:
	go run ./cmd/dltl -f td_api.tl -o _schema/tdapi.tl
.PHONY: download_tdlib_schema

check_generated: generate
	git diff --exit-code
.PHONY: check_generated
