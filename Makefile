.PHONY: help all install update dist publish
SHELL=/bin/bash

help: ## Print this message
	@awk 'BEGIN { FS = ":.*##"; print "Usage:  make <target>\n\nTargets:" } \
		/^[-_[:alpha:]]+:.?*##/ { printf "  %-15s%s\n", $$1, $$2 }' $(MAKEFILE_LIST)

all: ## Build binary
	dep ensure
	GOOS=linux GOARCH=amd64 go build

install: all ## Install binary
	GOOS=linux GOARCH=amd64 go install

update: ## Update dependencies
	dep ensure -update

dist: all ## Create Lambda distribution
	zip mario.zip ./mario-powerup

publish: dist ## Push the Lambda distribution to S3
	aws s3 cp mario.zip s3://mario-stage-lambda/mario.zip
	aws --region us-east-1 lambda update-function-code --function-name \
		mario-stage-lambda --s3-bucket mario-stage-lambda --s3-key mario.zip
