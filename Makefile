include ./linting.mk

.PHONY: ci
ci: lint unit_test

.PHONY: deps
deps:
	go mod tidy -v
	go mod download
	go mod vendor -v
	go mod verify

.PHONY: unit_test
unit_test:
	go test -v -cover ./... -count=1
