package main

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	apigateway "github.com/aws/aws-cdk-go/awscdk/v2/awsapigateway"
	ddb "github.com/aws/aws-cdk-go/awscdk/v2/awsdynamodb"
	lambda "github.com/aws/aws-cdk-go/awscdk/v2/awslambda"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

type Infra struct {
	scope constructs.Construct
	env   string
}

func NewInfra(scope constructs.Construct, env string) *Infra {
	return &Infra{
		scope: scope,
		env:   env,
	}
}

func (i *Infra) CreateDynamoTable(tableName string, partitionKey, sortKey *ddb.Attribute) ddb.Table {
	fullTableName := tableName + i.env
	props := &ddb.TableProps{
		TableName:     jsii.String(fullTableName),
		PartitionKey:  partitionKey,
		BillingMode:   ddb.BillingMode_PAY_PER_REQUEST,
		RemovalPolicy: awscdk.RemovalPolicy_DESTROY,
	}

	if sortKey != nil {
		props.SortKey = sortKey
	}

	return ddb.NewTable(i.scope, jsii.String(fullTableName), props)
}

func (i *Infra) CreateLambda(functionName string) lambda.Function {
	fullFunctionName := functionName + i.env
	return lambda.NewFunction(i.scope, jsii.String(fullFunctionName), &lambda.FunctionProps{
		FunctionName: jsii.String(fullFunctionName),
		Runtime:      lambda.Runtime_GO_1_X(),
		Handler:      jsii.String("main"),
		Environment: &map[string]*string{
			"ENVIRONMENT": jsii.String(i.env),
		},
	})
}

func (i *Infra) CreateApiGateway(apiName string, lambdaFn lambda.Function) apigateway.RestApi {
	fullApiName := apiName + i.env
	api := apigateway.NewRestApi(i.scope, jsii.String(fullApiName), &apigateway.RestApiProps{
		RestApiName: jsii.String(fullApiName),
		DeployOptions: &apigateway.StageOptions{
			StageName: jsii.String(i.env),
		},
	})

	authorizer := apigateway.NewTokenAuthorizer(i.scope, jsii.String(fullApiName+"Authorizer"), &apigateway.TokenAuthorizerProps{
		Handler: lambdaFn,
	})

	api.Root().AddMethod(jsii.String("ANY"), apigateway.NewLambdaIntegration(lambdaFn, &apigateway.LambdaIntegrationOptions{}), &apigateway.MethodOptions{
		AuthorizationType: apigateway.AuthorizationType_CUSTOM,
		Authorizer:        authorizer,
	})

	return api
}
