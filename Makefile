-include Makefile.local

REGISTRY ?=
IMAGE_NAME ?= song-guessing-game
TAG ?= latest
PLATFORM ?= linux/amd64
IMAGE := $(if $(strip $(REGISTRY)),$(REGISTRY)/,)$(IMAGE_NAME):$(TAG)

.PHONY: all build push

all: push

build:
	docker buildx build --platform $(PLATFORM) --tag $(IMAGE) --load .

push:
	@if [ -z "$(REGISTRY)" ]; then \
		printf '%s\n' 'Set REGISTRY or create an ignored Makefile.local before pushing.' >&2; \
		exit 1; \
	fi
	docker buildx build --platform $(PLATFORM) --tag $(IMAGE) --push .
