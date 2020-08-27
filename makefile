MAKEFLAGS += --warn-undefined-variables
SHELL := bash
.SHELLFLAGS := -eu -o pipefail -c
.DEFAULT_GOAL := test
.DELETE_ON_ERROR:
.SUFFIXES:

.PHONY: test
test: unit_test

.PHONY: unit_test
unit_test:
	go test ./... -cover -race

