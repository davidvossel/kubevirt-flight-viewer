REGISTRY ?= quay.io/dvossel
TAG ?= latest
IMAGE_NAME ?= kubevirt-flight-viewer

.PHONY: build
build:
	go build -o kubevirt-flight-viewer main.go

.PHONY: docker-build
docker-build:
	docker build . -t $(IMAGE_NAME):$(TAG) --file Dockerfile

.PHONY: docker-push
docker-push:
	docker push $(IMAGE_NAME):$(TAG)


.PHONY: test
test:
	go test
