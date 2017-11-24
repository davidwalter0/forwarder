# To enable kubernetes commands a valid configuration is required
export kubectl=${GOPATH}/bin/kubectl  --kubeconfig=${PWD}/cluster/auth/kubeconfig
export GOPATH=/go
SHELL=/bin/bash
MAKEFILE_DIR := $(patsubst %/,%,$(dir $(abspath $(lastword $(MAKEFILE_LIST)))))
CURRENT_DIR := $(notdir $(patsubst %/,%,$(dir $(MAKEFILE_DIR))))
DIR=$(MAKEFILE_DIR)

.PHONY: deps install clean image build push

# depends:=$(wildcard listener/*.go) $(wildcard kubeconfig/*.go) $(wildcard set/*.go) $(wildcard tracer/*.go) $(wildcard mgr/*.go)
depends:=$(wildcard listener/*.go kubeconfig/*.go set/*.go tracer/*.go mgr/*.go)

build_deps:=$(wildcard *.go)
target:=bin/$(notdir $(PWD))

all: $(target) 
	@echo Target $(target)
	@echo Build deps $(build_deps)
	@echo Depends $(depends)

# git:
# 	git add $(depends) $(build_deps) Makefile pipes.yaml daemonset.yaml.tmpl Dockerfile .version

etags:
	etags $(depends) $(build_deps)

.dep: $(target) $(depends) Makefile
	touch .dep

$(target): $(build_deps) $(depends) Makefile
	@echo "Building via % rule for $@ from $<"
	@echo $(depends)
	@if go version|grep -q 1.4 ; then											\
	    args="-s -w -X main.Version $${version} -X main.Build $$(date -u +%Y.%m.%d.%H.%M.%S.%:::z) -X main.Commit $$(git log --format=%h-%aI -n1)";	\
	fi;															\
	if go version|grep -qE "(1\.[5-9](\.?[0-9])*|1.[1-9][0-9](\.?[0-9])+|2.[0-9](\.?[0-9])*)"; then								\
            . .version;																		\
	    args="-s -w -X main.Version=$${version} -X main.Build=$$(date -u +%Y.%m.%d.%H.%M.%S.%:::z) -X main.Commit=$$(git log --format=%h-%aI -n1)";	\
	fi;															\
	CGO_ENABLED=0 go build --tags netgo -ldflags "$${args}" -o $@ $(build_deps) ;

install: build
	cp $(target) /go/bin/

image: build
	docker build --tag=davidwalter/$(notdir $(PWD)):latest .
	. .version; \
	docker tag davidwalter/$(notdir $(PWD)):latest \
	davidwalter/$(notdir $(PWD)):$${version}

push: image
	docker push davidwalter/$(notdir $(PWD)):latest
	. .version; \
	docker push davidwalter/$(notdir $(PWD)):$${version}

yaml: build
	applytmpl < daemonset.yaml.tmpl > daemonset.yaml

delete: yaml
	$(kubectl) delete ds/forwarder || true

# apply: yaml delete
apply: yaml
	$(kubectl) apply -f daemonset.yaml

clean:
	@if [[ -x "$(target)" ]]; then rm -f $(target); fi
	@if [[ -d "bin" ]]; then rmdir bin; fi
