# Makefile for the go target

# Golang
GO_ENV := GOCACHE=$(CURDIR)/.gocache
GO = GO111MODULE=on $(GO_ENV) go

.PHONY: \
	help

## Default
help: ## List available make targets with descriptions.
	@printf "\nuv is a hard dependency; Install from https://docs.astral.sh/uv/ if you haven't done so already.\n\n"
	@printf "Available targets:\n"
	@grep -hE '.*##\s' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*##"} {printf "  %-16s %s\n", $$1, $$2}'

## Building & Packaging
pack: ## Package Go binaries for installation/development test
	@if [ -f exp/main.go ]; then \
		$(GO) build -o dist/exp ./exp; \
	else \
		echo "Warning: no sources; skipping build."; \
	fi
