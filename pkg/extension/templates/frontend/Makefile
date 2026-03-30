REGISTRY ?= docker.io
NAMESPACE ?= kubespheredev
NAME ?= [[ .Name ]]
TAG ?= latest
TAG := $(or $(TAG),latest)
IMAGE ?= $(REGISTRY)/$(NAMESPACE)/$(NAME)-frontend

.PHONY: build build-assets push

# Image builds iframe assets inside Docker (see Dockerfile). Use build-assets for local-only output under dist/assets/.
build:
	docker build -t $(IMAGE):$(TAG) .

build-assets:
	npm install && npm run build:assets

push:
	docker push $(IMAGE):$(TAG)
