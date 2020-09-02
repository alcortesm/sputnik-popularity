MAKEFLAGS += --warn-undefined-variables
SHELL := bash
.SHELLFLAGS := -eu -o pipefail -c
.DEFAULT_GOAL := test
.DELETE_ON_ERROR:
.SUFFIXES:

.PHONY: test
test: unit integration

.PHONY: unit
unit:
	cd app && go test ./... -cover -race

.PHONY: clean
clean: integration_clean

.PHONY: integration
integration: integration_test integration_clean

integration_test: dir := tests/integration
integration_test: project := integration
integration_test: dc-flags := -p $(project) -f $(dir)/docker-compose.yml
integration_test:
	docker-compose $(dc-flags) build
	docker-compose $(dc-flags) up \
		--force-recreate \
		--remove-orphans \
		--abort-on-container-exit \
		--exit-code-from tester \

.PHONY: integration_clean
integration_clean: dir := tests/integration
integration_clean: project := integration
integration_clean: dc-flags := -p $(project) -f $(dir)/docker-compose.yml
integration_clean:
	docker-compose $(dc-flags) rm -v --stop --force
