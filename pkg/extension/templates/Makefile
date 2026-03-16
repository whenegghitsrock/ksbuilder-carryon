REGISTRY ?= docker.io
NAMESPACE ?= myorg
EXT_NAME ?= [[ .Name ]]

.PHONY: build build-frontend build-backend push push-frontend push-backend
build: build-frontend build-backend
push: push-frontend push-backend

build-frontend:
[[ if .HasFrontend ]]
	$(MAKE) -C frontend build NAME=$(EXT_NAME) REGISTRY=$(REGISTRY) NAMESPACE=$(NAMESPACE)
[[ else ]]
	@echo "Frontend not enabled, skip."
[[ end ]]

build-backend:
[[ if .HasBackend ]]
	$(MAKE) -C backend build NAME=$(EXT_NAME) REGISTRY=$(REGISTRY) NAMESPACE=$(NAMESPACE)
[[ else ]]
	@echo "Backend not enabled, skip."
[[ end ]]

push-frontend:
[[ if .HasFrontend ]]
	$(MAKE) -C frontend push NAME=$(EXT_NAME) REGISTRY=$(REGISTRY) NAMESPACE=$(NAMESPACE) TAG=$(TAG)
[[ else ]]
	@echo "Frontend not enabled, skip."
[[ end ]]

push-backend:
[[ if .HasBackend ]]
	$(MAKE) -C backend push NAME=$(EXT_NAME) REGISTRY=$(REGISTRY) NAMESPACE=$(NAMESPACE) TAG=$(TAG)
[[ else ]]
	@echo "Backend not enabled, skip."
[[ end ]]
