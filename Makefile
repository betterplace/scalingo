GOPATH := $(shell echo `pwd`/gospace)

.EXPORT_ALL_VARIABLES:

.PHONY: build build-info

default: fetch

clobber:
	rm -vrf $(GOPATH)/src/*

fetch:
	go get -u github.com/antonholmquist/jason
	go get -u github.com/stretchr/testify

reset: clobber fetch

test:
	@go test

coverage:
	@go test -coverprofile=coverage.out

coverage-display: coverage
	@go tool cover -html=coverage.out

tags:
	@gotags -tag-relative=false -silent=true -R=true -f $@ . $(GOPATH)
