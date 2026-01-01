.PHONY: fmt lint test check build cover snap node-deps playwright-install run start gifgrep termcaps-e2e

GIFGREP_ARGS ?=
BINDIR ?= bin
GIFGREP_BIN := $(BINDIR)/gifgrep
GIFGREP_DEPS := $(shell git ls-files '*.go' go.mod go.sum internal/assets/*.png)

# Allow: `make gifgrep tui skynet` (extra make goals become args).
ifneq (,$(filter $(firstword $(MAKECMDGOALS)),gifgrep run start))
  ifeq (,$(GIFGREP_ARGS))
    GIFGREP_ARGS := $(wordlist 2,$(words $(MAKECMDGOALS)),$(MAKECMDGOALS))
  endif
  $(eval $(GIFGREP_ARGS):;@:)
endif

fmt:
	gofumpt -w .

lint:
	golangci-lint run

test:
	go test ./... -cover

check:
	go vet ./...

build:
	go build ./...

cover:
	go test ./... -coverprofile=coverage.out
	go tool cover -func=coverage.out

snap:
	node scripts/ghostty-web-snap.mjs

termcaps-e2e:
	bash scripts/termcaps-e2e-macos.sh

node-deps:
	npm install

playwright-install:
	npx playwright install chromium

$(GIFGREP_BIN): $(GIFGREP_DEPS)
	@mkdir -p $(BINDIR)
	go build -o $(GIFGREP_BIN) ./cmd/gifgrep

gifgrep run start: $(GIFGREP_BIN)
	$(GIFGREP_BIN) $(GIFGREP_ARGS)
