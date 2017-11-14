
# Makefile to create and deploy cdn-securitygroup-sync as AWS lambda function

# These arguments are required for successful S3 upload and stack deployment
AWS_REGION ?= eu-west-1
AWS_ACCOUNT_ID ?=
SSM_KEY_ID ?=
S3_BUCKET ?=
SSM_KEY_PREFIX ?= css

# Leave as-is
VERSION ?= $(shell git describe --tags | cut -dv -f2)
S3_KEY ?= code/cdn-securitygroup-sync-$(VERSION).zip
LDFLAGS := -X main.AppVersion=$(VERSION) -w
# From: https://raw.githubusercontent.com/eawsy/aws-lambda-go-shim/master/src/Makefile.example
HANDLER ?= handler
PACKAGE ?= $(HANDLER)
MAKEFILE = $(word $(words $(MAKEFILE_LIST)),$(MAKEFILE_LIST))
CURDIR := $(shell pwd)
CF_TEMPLATE := cdn-securitygroup-sync.yaml
FN_NAME ?= $(SSM_KEY_PREFIX)-cdn-securitygroup-sync
# for binary release creation
PLATFORMS ?= linux/amd64 darwin/amd64 windows/amd64

all: dependencies docker upload

zip: build pack perm

dependencies:
	# fetch build dependencies
	go get -v
	docker pull eawsy/aws-lambda-go-shim:latest
	touch dependencies

build:
	# compile (inside docker)
	go build -buildmode=plugin -ldflags='-w -s $(LDFLAGS)' -o $(HANDLER).so

deploy-prebuilt:
	# deploy stack
	aws cloudformation --region $(AWS_REGION) deploy \
		--parameter-overrides FunctionName=$(FN_NAME) S3Bucket=$(S3_BUCKET) S3Key=$(S3_KEY) AccountId=$(AWS_ACCOUNT_ID) SSMSource=$(SSM_KEY_PREFIX) KeyId=$(SSM_KEY_ID) Region=$(AWS_REGION) \
		--template-file $(CF_TEMPLATE) --stack-name $(FN_NAME) \
		--capabilities CAPABILITY_IAM CAPABILITY_NAMED_IAM

deploy-source: dependencies docker upload
	# deploy stack
	aws cloudformation --region $(AWS_REGION) deploy \
		--parameter-overrides FunctionName=$(FN_NAME) S3Bucket=$(S3_BUCKET) S3Key=$(S3_KEY) AccountId=$(AWS_ACCOUNT_ID) SSMSource=$(SSM_KEY_PREFIX) KeyId=$(SSM_KEY_ID) Region=$(AWS_REGION) \
		--template-file $(CF_TEMPLATE) --stack-name $(FN_NAME) \
		--capabilities CAPABILITY_IAM CAPABILITY_NAMED_IAM

docker:
	# GOPATH must be defined!
	docker run --rm\
		-e HANDLER=$(HANDLER)\
		-e PACKAGE=$(PACKAGE)\
		-e LDFLAGS='$(LDFLAGS)'\
		-e GOPATH=$(GOPATH)\
		-v $(CURDIR):$(CURDIR)\
		$(foreach GP,$(subst :, ,$(GOPATH)),-v $(GP):$(GP))\
		-w $(CURDIR)\
		eawsy/aws-lambda-go-shim:latest make -f $(MAKEFILE) zip VERSION=$(VERSION)

pack:
	# create lambda zip
	pack $(HANDLER) $(HANDLER).so $(PACKAGE).zip

perm:
	chown $(shell stat -c '%u:%g' .) $(HANDLER).so $(PACKAGE).zip

run:
	aws lambda invoke                                                            \
  --function-name $(FN_NAME)                                                   \
  --invocation-type RequestResponse                                            \
  --log-type Tail  /dev/stderr                                                 \
  --query 'LogResult'                                                          \
  --output text |                                                              \
  base64 -D #-d fixme mac vs linux

upload:
	# upload lambda zip to S3
	aws s3 --region=$(AWS_REGION) cp $(PACKAGE).zip s3://$(S3_BUCKET)/$(S3_KEY)

release: dependencies docker
	mv $(PACKAGE).zip cdn-securitygroup-sync-lambda-$(VERSION).zip
	env GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS) -s" -o cdn-securitygroup-sync-linux-amd64
	env GOOS=darwin GOARCH=amd64 go build -ldflags "$(LDFLAGS) -s" -o cdn-securitygroup-sync-darwin-amd64
	env GOOS=windows GOARCH=amd64 go build -ldflags "$(LDFLAGS) -s" -o cdn-securitygroup-sync-windows-amd64
	for bin in cdn-securitygroup-sync-darwin* cdn-securitygroup-sync-linux* cdn-securitygroup-sync-windows*; \
		do zip $${bin}_v$(VERSION).zip $$bin; \
	done

clean:
	rm -rf $(HANDLER).so $(PACKAGE).zip dependencies cdn-securitygroup-sync-*
