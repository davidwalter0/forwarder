

.PHONY: deps install clean image build push
export GOPATH=/go
SHELL=/bin/bash
MAKEFILE_DIR := $(patsubst %/,%,$(dir $(abspath $(lastword $(MAKEFILE_LIST)))))
CURRENT_DIR := $(notdir $(patsubst %/,%,$(dir $(MAKEFILE_DIR))))
DIR=$(MAKEFILE_DIR)

target:=bin/$(notdir $(PWD))
# target:=$(GOPATH)/bin/$(notdir $(PWD))

all: $(target)

$(target) : $(wildcard *.go)

build: $(target)

.dep: $(target) Makefile
	touch .dep

%: bin/%

bin/%: %.go 
	@echo "Building via % rule for $@ from $<"
	@if go version|grep -q 1.4 ; then											\
	    args="-s -w -X main.Build $$(date -u +%Y.%m.%d.%H.%M.%S.%:::z) -X main.Commit $$(git log --format=%hash-%aI -n1)";	\
	fi;															\
	if go version|grep -qE "(1\.[5-9](\.?[0-9])*|1.[1-9][0-9](\.?[0-9])+|2.[0-9](\.?[0-9])*)"; then				\
	    args="-s -w -X main.Build=$$(date -u +%Y.%m.%d.%H.%M.%S.%:::z) -X main.Commit=$$(git log --format=%hash-%aI -n1)";	\
	fi;															\
	CGO_ENABLED=0 go build --tags netgo -ldflags "$${args}" -o $@ $^ ;

install: build
	cp $(target) /go/bin/
image: build
	docker build --tag=davidwalter/$(notdir $(PWD)):latest .
push: image
	docker push davidwalter/$(notdir $(PWD)):latest
clean:
	@if [[ -x "$(target)" ]]; then rm -f $(target); fi
	@if [[ -d "bin" ]]; then rmdir bin; fi
