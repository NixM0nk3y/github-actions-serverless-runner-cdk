# Meta tasks
# ----------

# Useful variables

# github details
export GITHUB_APP_ID ?= 158737
export GITHUB_INSTALLATION_ID ?= 21302729
export GITHUB_APP_KEY_PATH ?= private-key.pem
export GITHUB_HOOK_SECRET ?= BBBB

# deployment environment
export TENANT ?= openenterprise
export ENVIRONMENT ?= development

# aws environment
export AWS_REGION ?= eu-west-1
export AWS_ACCOUNT ?= 074705540277

# Output helpers
# --------------

TASK_DONE = echo "✓  $@ done"
TASK_BUILD = echo "🛠️  $@ done"

# ----------------
STACKS = $(shell find ./cmd/ -mindepth 1 -maxdepth 1 -type d)

export CODEBUILD_BUILD_NUMBER ?= 0
export CODEBUILD_RESOLVED_SOURCE_VERSION ?=$(shell git rev-list -1 HEAD --abbrev-commit)
export BUILD_DATE=$(shell date -u '+%Y%m%d')

.DEFAULT_GOAL := build

test:
	go test -v -p 1 ./...
	@$(TASK_BUILD)

bootstrap:
	CDK_NEW_BOOTSTRAP=1 cdk bootstrap aws://$(AWS_ACCOUNT)/$(AWS_REGION)
	@$(TASK_BUILD)

diff: diff/application
	@$(TASK_DONE)

synth: synth/application
	@$(TASK_DONE)

deploy: deploy/application
	@$(TASK_DONE)

destroy: destroy/application
	@$(TASK_DONE)

security/scan: build
	go run ./security/security.go
	@$(TASK_BUILD)

synth/application: build
	cdk synth --app ./application
	@$(TASK_BUILD)

diff/application: build
	cdk diff --app ./application
	@$(TASK_BUILD)

deploy/application: build
	cdk deploy --app ./application
	@$(TASK_BUILD)

ci/deploy/application: build
	cdk deploy --app ./application --ci true --require-approval never 
	@$(TASK_BUILD)

destroy/application: 
	cdk destroy --app ./application
	@$(TASK_BUILD)

build: stacks/build
	@$(TASK_DONE)

.PHONY: stacks/build $(STACKS)

stacks/build: $(STACKS)
	@$(TASK_DONE)

$(STACKS):
	go build -v ./$@
	@$(TASK_BUILD)    
	
init: 
	go mod download
	@$(TASK_BUILD)

