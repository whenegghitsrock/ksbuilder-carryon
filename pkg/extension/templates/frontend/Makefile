REGISTRY ?= docker.io
NAMESPACE ?= myorg
NAME ?= [[ .Name ]]
IMAGE ?= $(REGISTRY)/$(NAMESPACE)/$(NAME)-frontend
TAG ?= latest

build:
	docker build -t $(IMAGE):$(TAG) .

push:
	docker push $(IMAGE):$(TAG)
