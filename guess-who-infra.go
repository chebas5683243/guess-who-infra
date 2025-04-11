package main

import (
	"os"

	"github.com/chebas5683243/guess-who-infra/config"
	"github.com/chebas5683243/guess-who-infra/environment"
	"github.com/chebas5683243/guess-who-infra/infra"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"github.com/joho/godotenv"
)

type GuessWhoInfraStackProps struct {
	awscdk.StackProps
}

func NewGuessWhoInfraStack(scope constructs.Construct, id string, props *GuessWhoInfraStackProps) awscdk.Stack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	developmentStack := infra.NewEnvironmentNestedStack(stack, &infra.EnvironmentNestedStackProps{
		Environment: environment.Development,
	})
	developmentStack.CreateInfraestructure()

	productionStack := infra.NewEnvironmentNestedStack(stack, &infra.EnvironmentNestedStackProps{
		Environment: environment.Production,
	})
	productionStack.CreateInfraestructure()

	pipelineStack := infra.NewPipelineNestedStack(stack, "PipelineStack", &infra.PipelineNestedStackProps{
		GithubConfig: infra.GithubConfig{
			Owner:  os.Getenv("GITHUB_OWNER"),
			Repo:   os.Getenv("GITHUB_REPO"),
			Token:  os.Getenv("GITHUB_TOKEN"),
			Branch: os.Getenv("GITHUB_BRANCH"),
		},
		Stacks: infra.Stacks{
			Development: developmentStack,
			Production:  productionStack,
		},
	})
	pipelineStack.CreateInfraestructure()

	return stack
}

func main() {
	defer jsii.Close()

	if err := godotenv.Load(); err != nil {
		panic("Error loading .env file")
	}

	app := awscdk.NewApp(nil)

	NewGuessWhoInfraStack(app, config.StackName, &GuessWhoInfraStackProps{
		awscdk.StackProps{
			Env:       env(),
			StackName: jsii.String(config.StackName),
		},
	})

	app.Synth(nil)
}

func env() *awscdk.Environment {
	return &awscdk.Environment{
		Account: jsii.String(os.Getenv("AWS_CDK_DEFAULT_ACCOUNT")),
		Region:  jsii.String(os.Getenv("AWS_CDK_DEFAULT_REGION")),
	}
}
