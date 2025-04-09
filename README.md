# Guess Who Infrastructure

This repository contains the AWS CDK infrastructure code for the Guess Who game application.

## Architecture

The infrastructure consists of:

- API Gateway with a token authorizer
- Single Lambda function that handles both API requests and authorization
- Three DynamoDB tables:
  - Users (PK: id)
  - Game (PK: id)
  - Player (PK: gameId, SK: id)
- CI/CD Pipeline for automated deployments

## Prerequisites

- AWS CDK CLI installed
- AWS CLI configured with appropriate credentials
- Go 1.16 or later
- Node.js and npm

## Setup

1. Install AWS CDK:
```bash
npm install -g aws-cdk
```

2. Install Go dependencies:
```bash
go mod tidy
```

3. Configure your GitHub credentials:
   - Create a GitHub personal access token with `repo` scope
   - Store the token in AWS Secrets Manager
   - Update the `GitHubToken` value in `guess-who-infra.go` with your secret name

4. Update the GitHub configuration:
   - Update `GitHubOwner` and `GitHubRepo` in `guess-who-infra.go` with your repository details

## Deployment

1. Bootstrap your AWS environment (if not already done):
```bash
cdk bootstrap
```

2. Deploy the infrastructure:
```bash
cdk deploy --all
```

This will create:
- Development and Production stacks
- CI/CD Pipeline
- All necessary AWS resources

## CI/CD Pipeline

The pipeline is triggered on changes to the main branch and includes:

1. Source Stage: Pulls code from GitHub
2. Build Stage: Runs tests and builds the Lambda function
3. Deploy Dev Stage: Deploys to development environment
4. Approval Stage: Manual approval required
5. Deploy Prod Stage: Deploys to production environment

## Development

1. Make changes to your code in the main repository
2. Push changes to the main branch
3. The pipeline will automatically:
   - Run tests
   - Deploy to development
   - Wait for approval
   - Deploy to production (if approved)

## Infrastructure Structure

- `guess-who-infra.go`: Main stack definition
- `infrastructure.go`: Infrastructure components (tables, lambda, API Gateway)
- `pipeline.go`: CI/CD pipeline definition

## Environment Separation

Resources are separated by environment using suffixes:
- Development: `ResourceNameDev`
- Production: `ResourceNameProd`

## Security

- API Gateway uses a token authorizer
- Lambda has least-privilege access to DynamoDB tables
- GitHub token is stored in AWS Secrets Manager
