package main

import (
	"os"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsdynamodb"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"github.com/joho/godotenv"
)

const (
	AppName = "GuessWho"
)

type GuessWhoInfraStackProps struct {
	awscdk.StackProps
	Environment string
}

func NewGuessWhoInfraStack(scope constructs.Construct, id string, props *GuessWhoInfraStackProps) awscdk.Stack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	infra := NewInfra(stack, props.Environment)

	usersTable := infra.CreateDynamoTable(
		"Users",
		&awsdynamodb.Attribute{
			Name: jsii.String("id"),
			Type: awsdynamodb.AttributeType_STRING,
		},
		nil)

	gameTable := infra.CreateDynamoTable(
		"Game",
		&awsdynamodb.Attribute{
			Name: jsii.String("id"),
			Type: awsdynamodb.AttributeType_STRING,
		},
		nil)

	playerTable := infra.CreateDynamoTable(
		"Player",
		&awsdynamodb.Attribute{
			Name: jsii.String("gameId"),
			Type: awsdynamodb.AttributeType_STRING,
		},
		&awsdynamodb.Attribute{
			Name: jsii.String("id"),
			Type: awsdynamodb.AttributeType_STRING,
		})

	lambda := infra.CreateLambda(AppName + "Lambda")

	usersTable.GrantReadWriteData(lambda)
	gameTable.GrantReadWriteData(lambda)
	playerTable.GrantReadWriteData(lambda)

	infra.CreateApiGateway(AppName+"Api", lambda)

	return stack
}

func main() {
	defer jsii.Close()

	if err := godotenv.Load(); err != nil {
		panic("Error loading .env file")
	}

	app := awscdk.NewApp(nil)

	NewGuessWhoInfraStack(app, AppName+"InfraStackDev", &GuessWhoInfraStackProps{
		StackProps: awscdk.StackProps{
			Env: env(),
		},
		Environment: "Dev",
	})

	NewGuessWhoInfraStack(app, AppName+"InfraStackProd", &GuessWhoInfraStackProps{
		StackProps: awscdk.StackProps{
			Env: env(),
		},
		Environment: "Prod",
	})

	NewPipelineStack(app, AppName+"PipelineStack", &PipelineStackProps{
		StackProps: awscdk.StackProps{
			Env: env(),
		},
		GitHubOwner: os.Getenv("GITHUB_OWNER"),
		GitHubRepo:  os.Getenv("GITHUB_REPO"),
	})

	app.Synth(nil)
}

func env() *awscdk.Environment {
	return &awscdk.Environment{
		Account: jsii.String(os.Getenv("AWS_CDK_DEFAULT_ACCOUNT")),
		Region:  jsii.String(os.Getenv("AWS_CDK_DEFAULT_REGION")),
	}
}
