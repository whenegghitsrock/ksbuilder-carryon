REGISTRY ?= docker.io
NAMESPACE ?= kubespheredev
NAME ?= [[ .Name ]]
TAG ?= latest
TAG := $(or $(TAG),latest)
IMAGE ?= $(REGISTRY)/$(NAMESPACE)/$(NAME)-frontend

build:
	docker build -t $(IMAGE):$(TAG) .

push:
	docker push $(IMAGE):$(TAG)
