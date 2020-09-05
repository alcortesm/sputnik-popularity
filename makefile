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
clean: integration_clean e2e_clean

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
		--detach
	result=$$(docker wait $(project)_tester_1); \
	if [ $${result} != 0 ]; then \
		docker-compose $(dc-flags) stop; \
		docker-compose $(dc-flags) logs; \
		false; \
	fi
	docker-compose $(dc-flags) logs tester

.PHONY: integration_clean
integration_clean: dir := tests/integration
integration_clean: project := integration
integration_clean: dc-flags := -p $(project) -f $(dir)/docker-compose.yml
integration_clean:
	docker-compose $(dc-flags) rm -v --stop --force

.PHONY: e2e
e2e: e2e_test e2e_clean

.PHONY: e2e_test
e2e_test: dir := tests/e2e
e2e_test: project := e2e
e2e_test: dc-flags := -p $(project) -f $(dir)/docker-compose.yml
e2e_test:
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

.PHONY: e2e_clean
e2e_clean: dir := tests/e2e
e2e_clean: project := e2e
e2e_clean: dc-flags := -p $(project) -f $(dir)/docker-compose.yml
e2e_clean:
	docker-compose $(dc-flags) rm -v --stop --force
