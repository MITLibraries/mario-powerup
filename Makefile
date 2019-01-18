.PHONY: help all install update dist publish
SHELL=/bin/bash
ECR_REGISTRY=672626379771.dkr.ecr.us-east-1.amazonaws.com

help: ## Print this message
	@awk 'BEGIN { FS = ":.*##"; print "Usage:  make <target>\n\nTargets:" } \
		/^[-_[:alpha:]]+:.?*##/ { printf "  %-15s%s\n", $$1, $$2 }' $(MAKEFILE_LIST)

all: ## Build binary
	dep ensure
	GOOS=linux go build

install: all ## Install binary
	GOOS=linux go install

update: ## Update dependencies
	dep ensure -update

dist: all ## Create Lambda distribution
	zip mario.zip ./mario-powerup

publish: dist ## Push the Lambda distribution to S3
	$$(aws s3 cp mario.zip s3://mario-staging-lambda/mario.zip)
