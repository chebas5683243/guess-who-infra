.PHONY: bootstrap build lambda-build deploy clean setup

# Default target
all: setup build lambda-build

# Setup permissions and dependencies
setup:
	chmod +x base-lambda/build.sh

# Bootstrap CDK
bootstrap: setup
	cdk bootstrap

# Build the main CDK application
build:
	go mod tidy
	go build

# Build the Lambda function
lambda-build: setup
	cd base-lambda && \
	go mod tidy && \
	./build.sh

# Deploy the stack
deploy: build lambda-build
	cdk deploy --all

# Clean build artifacts
clean:
	rm -f guess-who-infra
	cd base-lambda && rm -f bootstrap

# Help target
help:
	@echo "Available targets:"
	@echo "  setup       - Setup permissions and dependencies"
	@echo "  bootstrap   - Bootstrap CDK environment"
	@echo "  build      - Build the main CDK application"
	@echo "  lambda-build - Build the Lambda function"
	@echo "  deploy     - Deploy the stack"
	@echo "  clean      - Clean build artifacts"
	@echo "  help       - Show this help message" 