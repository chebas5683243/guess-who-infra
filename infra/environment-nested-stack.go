package infra

import (
	"github.com/chebas5683243/guess-who-infra/config"
	"github.com/chebas5683243/guess-who-infra/customconstructs"
	"github.com/chebas5683243/guess-who-infra/environment"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsapigateway"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsdynamodb"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambda"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

type EnvironmentNestedStackProps struct {
	awscdk.NestedStackProps
	environment.Environment
}

type EnvironmentNestedStack struct {
	awscdk.NestedStack
	env    environment.Environment
	Lambda awslambda.Function
}

func NewEnvironmentNestedStack(scope constructs.Construct, props *EnvironmentNestedStackProps) *EnvironmentNestedStack {
	nestedStack := &EnvironmentNestedStack{}

	if props == nil {
		panic("environment nested stack props missing")
	}

	stackId := string(props.Environment) + "NestedStack"

	awscdk.NewNestedStack_Override(nestedStack, scope, &stackId, &props.NestedStackProps)
	nestedStack.env = props.Environment

	return nestedStack
}

func (nestedStack *EnvironmentNestedStack) CreateInfraestructure() {
	nestedStack.createLambdaFunction()
	nestedStack.createDynamoTables()
	nestedStack.createApiGateway()
}

func (nestedStack *EnvironmentNestedStack) createLambdaFunction() {
	lambdaConstruct := customconstructs.NewLambdaConstruct(
		nestedStack,
		"LambdasConstruct",
		&customconstructs.LambdaConstructProps{
			Environment: nestedStack.env,
		})

	nestedStack.Lambda = lambdaConstruct.CreateGoFunction(&customconstructs.LambdaFunctionProps{
		FunctionName: jsii.String("Game"),
		EnvVariables: &map[string]*string{
			"ENVIRONMENT": jsii.String(string(nestedStack.env)),
		},
	})
}

func (nestedStack *EnvironmentNestedStack) createDynamoTables() {
	ddbConstruct := customconstructs.NewDynamoDBConstruct(
		nestedStack,
		"DynamoConstruct",
		&customconstructs.DynamoDbConstructProps{
			Environment: nestedStack.env,
		},
	)

	usersTable := ddbConstruct.CreateTable(&customconstructs.DynamoDBTableProps{
		TableName: jsii.String("Users"),
		PartitionKey: &awsdynamodb.Attribute{
			Name: jsii.String("id"),
			Type: awsdynamodb.AttributeType_STRING,
		},
	})

	gamesTable := ddbConstruct.CreateTable(&customconstructs.DynamoDBTableProps{
		TableName: jsii.String("Games"),
		PartitionKey: &awsdynamodb.Attribute{
			Name: jsii.String("id"),
			Type: awsdynamodb.AttributeType_STRING,
		},
	})

	playersTable := ddbConstruct.CreateTable(&customconstructs.DynamoDBTableProps{
		TableName: jsii.String("Players"),
		PartitionKey: &awsdynamodb.Attribute{
			Name: jsii.String("gameId"),
			Type: awsdynamodb.AttributeType_STRING,
		},
		SortKey: &awsdynamodb.Attribute{
			Name: jsii.String("id"),
			Type: awsdynamodb.AttributeType_STRING,
		},
	})

	ddbConstruct.GrantAccessToLambda(usersTable, nestedStack.Lambda)
	ddbConstruct.GrantAccessToLambda(gamesTable, nestedStack.Lambda)
	ddbConstruct.GrantAccessToLambda(playersTable, nestedStack.Lambda)
}

func (nestedStack *EnvironmentNestedStack) createApiGateway() {
	fullApiName := config.StackName + "ApiGateway" + string(nestedStack.env)
	api := awsapigateway.NewRestApi(nestedStack, &fullApiName, &awsapigateway.RestApiProps{
		RestApiName: &fullApiName,
		DeployOptions: &awsapigateway.StageOptions{
			StageName: jsii.String(string(nestedStack.env)),
		},
	})

	healthResource := api.Root().AddResource(jsii.String("health"), &awsapigateway.ResourceOptions{})

	healthResource.AddMethod(
		jsii.String("GET"),
		awsapigateway.NewLambdaIntegration(nestedStack.Lambda, &awsapigateway.LambdaIntegrationOptions{}),
		&awsapigateway.MethodOptions{},
	)
	healthResource.AddMethod(
		jsii.String("POST"),
		awsapigateway.NewLambdaIntegration(nestedStack.Lambda, &awsapigateway.LambdaIntegrationOptions{}),
		&awsapigateway.MethodOptions{},
	)
}
