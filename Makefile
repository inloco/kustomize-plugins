.DEFAULT_GOAL = build
SHELL = /bin/bash

RESET := $(shell tput sgr0)
BOLD := $(shell tput bold)
RED := $(shell tput setaf 1)
EOL := \n

API_GROUP ?= incognia.com
API_VERSION ?= v1alpha1
PLACEMENT ?= $(shell echo $${XDG_CONFIG_HOME:-$$HOME/.config}/kustomize/plugin/${API_GROUP}/${API_VERSION})

setup-environment:
	@printf '${BOLD}${RED}make: *** [setup-environment]${RESET}${EOL}'
	$(eval SRC_PATH := $(shell pwd))
	$(eval TMP_PATH := $(shell mktemp -d))
	$(eval GIT_PATH := $(shell go list -m))
	$(eval MOD_PATH := ${TMP_PATH}/src/${GIT_PATH})
	$(eval VER_DESC := $(shell git describe --tags))
	export GOPATH='${TMP_PATH}'
	export GO111MODULE='on'
	mkdir -p ${MOD_PATH}
	rmdir ${MOD_PATH}
	ln -Fs ${SRC_PATH} ${MOD_PATH}
.PHONY: setup-environment

build: namespace/plugin unnamespaced/plugin clusterroles/plugin
.PHONY: build

namespace/plugin: setup-environment
	@printf '${BOLD}${RED}make: *** [namespace/plugin]${RESET}${EOL}'
	cd ${MOD_PATH}                              && \
	go build                                       \
		-o 'namespace/plugin'                      \
		-a                                         \
		-installsuffix 'cgo'                       \
		-gcflags 'all=-trimpath "${TMP_PATH}/src"' \
		-v 										   \
		./namespace

unnamespaced/plugin: setup-environment
	@printf '${BOLD}${RED}make: *** [unnamespaced/plugin]${RESET}${EOL}'
	cd ${MOD_PATH}                              && \
	go build                                       \
		-o 'unnamespaced/plugin'                   \
		-a                                         \
		-installsuffix 'cgo'                       \
		-gcflags 'all=-trimpath "${TMP_PATH}/src"' \
		-v 										   \
		./unnamespaced

clusterroles/plugin: setup-environment
	@printf '${BOLD}${RED}make: *** [clusterroles/plugin]${RESET}${EOL}'
	cd ${MOD_PATH}                              && \
	go build                                       \
		-o 'clusterroles/plugin'                   \
		-a                                         \
		-installsuffix 'cgo'                       \
		-gcflags 'all=-trimpath "${TMP_PATH}/src"' \
		-v 										   \
		./clusterroles

test: setup-environment
	@printf '${BOLD}${RED}make: *** [test]${RESET}${EOL}'
	cd ${MOD_PATH} && \
	go test           \
		-v ./...
.PHONY: test

continuous-integration: test build
.PHONY: continuous-integration

install: build install-namespace install-unnamespaced install-clusterroles
.PHONY: install

install-namespace:
	@printf '${BOLD}${RED}make: *** [install-namespace]${RESET}${EOL}'
	mkdir -p ${PLACEMENT}/namespaced
	cp ./namespace/plugin ${PLACEMENT}/namespace/Namespace
.PHONY: install-namespace

install-unnamespaced: setup-environment unnamespaced/plugin
	@printf '${BOLD}${RED}make: *** [install-unnamespaced]${RESET}${EOL}'
	mkdir -p ${PLACEMENT}/unnamespaced
	cp ./unnamespaced/plugin ${PLACEMENT}/unnamespaced/Unnamespaced
.PHONY: install-unnamespaced

install-clusterroles: setup-environment clusterroles/plugin
	@printf '${BOLD}${RED}make: *** [install-clusterroles]${RESET}${EOL}'
	mkdir -p ${PLACEMENT}/clusterroles
	cp ./clusterroles/plugin ${PLACEMENT}/clusterroles/ClusterRoles
.PHONY: install-clusterroles
