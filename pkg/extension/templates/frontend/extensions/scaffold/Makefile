REGISTRY ?= docker.io
NAMESPACE ?= kubespheredev
NAME ?= [[ .Name ]]
TAG ?= latest
TAG := $(or $(TAG),latest)
IMAGE ?= $(REGISTRY)/$(NAMESPACE)/$(NAME)-frontend

.PHONY: build build-assets push

# Image builds shell + iframe assets inside Docker (see Dockerfile). build-assets fills dist/ locally.
build:
	docker build -t $(IMAGE):$(TAG) .

build-assets:
	npm install && npm run build:assets && npm run build:shell

push:
	docker push $(IMAGE):$(TAG)
