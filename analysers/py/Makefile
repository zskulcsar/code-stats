# Makefile for the python target

LOG_LEVEL := DEBUG
UV = @PYTHONPYCACHEPREFIX=$(CURDIR)/.pycache uv

.PHONY: \
	help \
	fmt-py \
	run

## Default
help: ## List available make targets with descriptions.
	@printf "\nuv is a hard dependency; Install from https://docs.astral.sh/uv/ if you haven't done so already.\n\n"
	@printf "Available targets:\n"
	@grep -hE '.*##\s' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*##"} {printf "  %-16s %s\n", $$1, $$2}'

## Environment management
.venv: ## Create the python virtual environment and download all dependencies
	$(UV) venv  --prompt python-source-analyser .venv
	$(UV) pip install --python .venv --upgrade pip
	$(UV) sync

## Development targets
fmt-py: .venv ## Run `ruff format` over Python sources
	$(UV) run ruff format -s

run: fmt-py ## Runs the code analyser with the PARAMS passed
	$(UV) run python -m main --log=${LOG_LEVEL} $(PARAMS)

uv: ## Runs the arbitrary CMD with uv. ie.: make uv CMD="run ruff format --help"
	$(UV) $(CMD)