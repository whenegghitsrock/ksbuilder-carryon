REGISTRY ?= docker.io
NAMESPACE ?= myorg
NAME ?= [[ .Name ]]
IMAGE ?= $(REGISTRY)/$(NAMESPACE)/$(NAME)-api
TAG ?= latest

build:
	go mod tidy
	go build -o server .
	docker build -t $(IMAGE):$(TAG) .

push:
	docker push $(IMAGE):$(TAG)
