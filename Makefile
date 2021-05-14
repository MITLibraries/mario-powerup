.PHONY: help all install update dist publish
SHELL=/bin/bash
AWS_REGION=us-east-1

help: ## Print this message
	@awk 'BEGIN { FS = ":.*##"; print "Usage:  make <target>\n\nTargets:" } \
		/^[-_[:alpha:]]+:.?*##/ { printf "  %-15s%s\n", $$1, $$2 }' $(MAKEFILE_LIST)

all: ## Build binary
	GOOS=linux GOARCH=amd64 go build

install: all ## Install binary
	GOOS=linux GOARCH=amd64 go install

update: ## Update dependencies
	go get -u

dist: all ## Create Lambda distribution
	zip mario.zip ./mario-powerup

publish: dist ## Push the Lambda distribution to S3
	aws --region $(AWS_REGION) s3 cp mario.zip s3://mario-stage-lambda/mario.zip
	aws --region $(AWS_REGION) lambda update-function-code --function-name \
		mario-stage-lambda --s3-bucket mario-stage-lambda --s3-key mario.zip

promote: ## Copy the Lambda deployment from staging to production
	aws --region $(AWS_REGION) s3 sync s3://mario-stage-lambda/mario.zip \
		s3://mario-prod-lambda/mario.zip
	aws --region $(AWS_REGION) lambda update-function-code --function-name \
		mario-prod-lambda --s3-bucket mario-prod-lambda --s3-key mario.zip
