REGISTRY ?= docker.io
NAMESPACE ?= kubespheredev
NAME ?= [[ .Name ]]
TAG ?= latest
TAG := $(or $(TAG),latest)
IMAGE ?= $(REGISTRY)/$(NAMESPACE)/$(NAME)-api

build:
	go mod tidy
	go build -o server .
	docker build -t $(IMAGE):$(TAG) .

push:
	docker push $(IMAGE):$(TAG)
