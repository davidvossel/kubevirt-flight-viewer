REGISTRY ?= quay.io/dvossel
TAG ?= latest
IMAGE_NAME ?= kubevirt-flight-viewer

.PHONY: build
build:
	go build .

.PHONY: docker-build
docker-build:
	docker build . -t $(REGISTRY)/$(IMAGE_NAME):$(TAG) --file Dockerfile

.PHONY: docker-push
docker-push:
	docker push $(REGISTRY)/$(IMAGE_NAME):$(TAG)

.PHONY: generate
generate:
	hack/update-codegen.sh

.PHONY: verify
verify:
	hack/verify-codegen.sh

.PHONY: test
test:
	go test
