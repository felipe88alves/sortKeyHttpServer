
# Image URL to use all building/pushing image targets
REPO ?= webservice
IMG ?= sortedurlstats
TAG ?= 0.1

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

DATASOURCEMETHOD ?= file
DOCKER_PORT ?= 80

##@ General

# The help target prints out all targets with their descriptions organized
# beneath their categories. The categories are represented by '##@' and the
# target descriptions by '##'. The awk commands is responsible for reading the
# entire set of makefiles included in this invocation, looking for lines of the
# file as xyz: ## something, and then pretty-format the target and help. Then,
# if there's a line with ##@ something, that gets pretty-printed as a category.
# More info on the usage of ANSI control characters for terminal formatting:
# https://en.wikipedia.org/wiki/ANSI_escape_code#SGR_parameters
# More info on the awk command:
# http://linuxcommand.org/lc3_adv_awk.php

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development

.PHONY: fmt
fmt: ## Run go fmt against code.
	go fmt ./...

.PHONY: vet
vet: ## Run go vet against code.
	go vet ./...

.PHONY: pre-test
pre-test: fmt vet

.PHONY: test-unit
test-unit: pre-test --unit

--unit:
	go test ./... -coverprofile cover.out

.PHONY: run
run: fmt vet ## Run the webservice from host.
	go run ./...

##@ Build

.PHONY: build
build: fmt vet ## Build go binary.
	go build -o bin/${IMG}

.PHONY: docker-build
docker-build: fmt vet test-unit ## Build docker image with the webservice.
	docker build -t ${REPO}/${IMG}:${TAG} .

.PHONY: kind-load-docker-image
kind-load-docker-image: --kind ## Push docker image with the webservice to kind cluster.
	$(KIND) load docker-image ${REPO}/${IMG}:${TAG}

##@ Deployment Infra

.PHONY: kind-create-cluster
kind-create-cluster: --kind ## Deploy Kind k8s cluster.
	$(KIND) create cluster --config=./resources/infra/kind-cluster.yaml

.PHONY: kind-delete-cluster
kind-delete-cluster: --kind ## Delete Kind k8s cluster.
	$(KIND) delete cluster

##@ Deployment App

.PHONY: deploy-kustomize
deploy-kustomize: --kustomize ## Deploy urlstats app to the K8s cluster specified in ~/.kube/config using Kustomize.
	$(KUSTOMIZE) build resources/kustomize | kubectl apply -f -

.PHONY: undeploy-kustomize
undeploy-kustomize: --kustomize ## Undeploy urlstats app to the K8s cluster specified in ~/.kube/config using Kustomize.
	$(KUSTOMIZE) build resources/kustomize | kubectl delete -f -

.PHONY: deploy-k8s
deploy-k8s: ## Deploy urlstats app to the K8s cluster specified in ~/.kube/config using yaml files.
	kubectl apply -f resources/k8s-manifests/urlstats-deployment.yaml

.PHONY: undeploy-k8s
undeploy-k8s: ## Undeploy urlstats app to the K8s cluster specified in ~/.kube/config using yaml files.
	kubectl delete -f resources/k8s-manifests/urlstats-deployment.yaml

.PHONY: deploy-docker-file
deploy-docker-file: ## Deploy urlstats app locally using docker.
	docker compose -f docker-compose-file.yml up

.PHONY: deploy-docker-http
deploy-docker-http: ## Deploy urlstats app locally using docker.
	docker compose -f docker-compose-http.yml up

.PHONY: deploy-bin
deploy-bin: build ## Deploy urlstats app locally using go binary.
	DATA_COLLECTION_METHOD=${DATASOURCEMETHOD} ./bin/${IMG}

.PHONY: undeploy-bin
undeploy-bin: ## Uneploy urlstats app locally using go binary.
	kubectl delete -f resources/k8s-manifests/urlstats-deployment.yaml

##@ E2E Deployment
.PHONY: all
all: test-unit docker-build kind-delete-cluster kind-create-cluster kind-load-docker-image deploy-kustomize ## build docker image, provision k8's kinD cluster and deploy urlstats webservice using kustomize

##@ Support/Troubleshoot
.PHONY: kind-list-loaded-images
kind-list-loaded-images: --kind ## List Docker images loaded to Kind k8s cluster.
	docker exec -it kind-control-plane crictl images

.PHONY: wsl2-start-docker-daemon ## Usefull if docker is installed locally in wsl2
wsl2-start-docker-daemon: ## Start docker daemon in wsl2.
	./resources/scripts/wsl2_start_docker_daemon.sh

##@ Cleanup

.PHONY: clean-bin
clean-bin: ## Removes the bin directory
	rm -rf ./bin

.PHONY: clean-docker-images
clean-docker-images: ## Removes all docker images
	docker rmi -f $(shell docker images -aq)

.PHONY: clean-docker-containers
clean-docker-containers: ## Removes all docker containers. WARNING: also removes running containers.
	docker rm -vf $(shell docker ps -aq)

##@ Build Dependencies

## Location to install dependencies to
LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

## Tool Binaries
KIND ?= $(LOCALBIN)/kind
KUSTOMIZE ?= $(LOCALBIN)/kustomize

## Tool Versions
KIND_VERSION ?= v0.17.0
KUSTOMIZE_VERSION ?= v4.5.7

--kind: $(KIND) ## Download kind locally if necessary.
$(KIND): $(LOCALBIN)
	test -s $(LOCALBIN)/kind || GOBIN=$(LOCALBIN) go install sigs.k8s.io/kind@${KIND_VERSION}

KUSTOMIZE_INSTALL_SCRIPT ?= "https://raw.githubusercontent.com/kubernetes-sigs/kustomize/master/hack/install_kustomize.sh"
--kustomize: $(KUSTOMIZE) ## Download kustomize locally if necessary.
$(KUSTOMIZE): $(LOCALBIN)
	test -s $(LOCALBIN)/kustomize || curl -s $(KUSTOMIZE_INSTALL_SCRIPT) | bash -s -- $(subst v,,$(KUSTOMIZE_VERSION)) $(LOCALBIN)
