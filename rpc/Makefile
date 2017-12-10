MAKEFILE_DIR := $(patsubst %/,%,$(dir $(abspath $(lastword $(MAKEFILE_LIST)))))
CURRENT_DIR := $(notdir $(patsubst %/,%,$(dir $(MAKEFILE_DIR))))
export DIR=$(MAKEFILE_DIR)


targets:=bin/forwarder-server bin/forwarder-client bin/forwarder-rpc
deps:=pipe/pipe.go
generated:=$(patsubst pipe/%.proto,pipe/%.pb.go,$(wildcard pipe/*.proto))

.PHONY: server.bin client.bin forwarder.bin

all: bin proto build
	echo $(MAKEFILE_DIR)
	echo $(CURRENT_DIR)
	echo $(DIR)

bin:
	mkdir -p bin

proto-gen: generate proto

proto: $(wildcard pipe/*.proto)

build: $(targets) Makefile
	echo $(targets)

generate: $(generated)

pipe/%.pb.go: pipe/%.proto
	protoc --proto_path=pipe --proto_path=/go/src --go_out=plugins=grpc:pipe $^

bin/%: %/%.go $(generated) $(deps) 
	go build -tags netgo -ldflags '-w -s' -o $@ $^

bin/forwarder-server: server/main/server.go $(generated) $(deps) Makefile $(wildcard server/impl/*.go)
	echo $@ $^
	. ../.version; \
	args="-s -w -X main.Version=$${version} -X main.Build=$$(date -u +%Y.%m.%d.%H.%M.%S.%:::z) -X main.Commit=$$(git log --format=%h-%aI -n1)";	\
	go build -tags netgo -ldflags "$${args}" -o $@ $<

bin/forwarder-client: client/main/client.go  $(generated) $(deps) Makefile $(wildcard client/impl/*.go)
	echo $@ $^
	. ../.version; \
	args="-s -w -X main.Version=$${version} -X main.Build=$$(date -u +%Y.%m.%d.%H.%M.%S.%:::z) -X main.Commit=$$(git log --format=%h-%aI -n1)";	\
	go build -tags netgo -ldflags "$${args}" -o $@ $<

bin/forwarder-rpc: forwarder/server.go $(generated) $(deps) Makefile $(wildcard server/impl/*.go)
	echo $@ $^
	. ../.version; \
	args="-s -w -X main.Version=$${version} -X main.Build=$$(date -u +%Y.%m.%d.%H.%M.%S.%:::z) -X main.Commit=$$(git log --format=%h-%aI -n1)";	\
	go build -tags netgo -ldflags "$${args}" -o $@ $<

clean:
	rm -f bin/*
