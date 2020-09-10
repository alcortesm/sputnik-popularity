MAKEFLAGS += --warn-undefined-variables
SHELL := bash
.SHELLFLAGS := -eu -o pipefail -c
.DEFAULT_GOAL := test
.DELETE_ON_ERROR:
.SUFFIXES:

.PHONY: test
test: unit integration e2e

.PHONY: unit
unit:
	go test ./... -cover -race

.PHONY: clean
clean: integration-clean e2e-clean

.PHONY: integration
integration: integration-test integration-clean

integration-test: dir := tests/integration
integration-test: project := integration
integration-test: dc-flags := -p $(project) -f $(dir)/docker-compose.yml
integration-test:
	docker-compose $(dc-flags) build
	docker-compose $(dc-flags) up \
		--force-recreate \
		--remove-orphans \
		--detach
	result=$$(docker wait $(project)_tester_1); \
		if [ $${result} != 0 ]; then \
			docker-compose $(dc-flags) stop; \
			docker-compose $(dc-flags) logs; \
			false; \
		fi
	docker-compose $(dc-flags) logs tester

.PHONY: integration-clean
integration-clean: dir := tests/integration
integration-clean: project := integration
integration-clean: dc-flags := -p $(project) -f $(dir)/docker-compose.yml
integration-clean:
	docker-compose $(dc-flags) rm -v --stop --force

.PHONY: e2e
e2e: e2e-test e2e-clean

.PHONY: e2e-test
e2e-test: dir := tests/e2e
e2e-test: project := e2e
e2e-test: dc-flags := -p $(project) -f $(dir)/docker-compose.yml
e2e-test:
	docker-compose $(dc-flags) build
	docker-compose $(dc-flags) up \
		--force-recreate \
		--remove-orphans \
		--detach
	result=$$(docker wait $(project)_tester_1); \
		if [ $${result} != 0 ]; then \
			docker-compose $(dc-flags) stop; \
			docker-compose $(dc-flags) logs; \
			false; \
		fi
	docker-compose $(dc-flags) logs tester

.PHONY: e2e-clean
e2e-clean: dir := tests/e2e
e2e-clean: project := e2e
e2e-clean: dc-flags := -p $(project) -f $(dir)/docker-compose.yml
e2e-clean:
	docker-compose $(dc-flags) rm -v --stop --force

.PHONY: docker-image
docker-image:
	docker build -t sputnik --target=run-app .
