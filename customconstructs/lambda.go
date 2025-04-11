package customconstructs

import (
	"github.com/chebas5683243/guess-who-infra/config"
	"github.com/chebas5683243/guess-who-infra/environment"

	"github.com/aws/aws-cdk-go/awscdk/v2/awslambda"
	"github.com/aws/aws-cdk-go/awscdk/v2/awss3assets"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

type LambdaConstructProps struct {
	environment.Environment
}

type LambdaConstruct struct {
	constructs.Construct
	env string
}

type LambdaFunctionProps struct {
	FunctionName *string
	EnvVariables *map[string]*string
}

func NewLambdaConstruct(scope constructs.Construct, id string, props *LambdaConstructProps) *LambdaConstruct {
	construct := &LambdaConstruct{}
	constructs.NewConstruct_Override(construct, scope, &id)

	if props == nil {
		panic("lambda construct props missing")
	}

	construct.env = string(props.Environment)
	return construct
}

func (l *LambdaConstruct) CreateGoFunction(props *LambdaFunctionProps) awslambda.Function {
	if props == nil {
		panic("lambda configuration missing")
	}

	if props.FunctionName == nil {
		panic("missing function name for lambda function")
	}

	functionName := config.StackName + "__" + *props.FunctionName + l.env

	return awslambda.NewFunction(l, &functionName, &awslambda.FunctionProps{
		FunctionName: &functionName,
		Runtime:      awslambda.Runtime_PROVIDED_AL2023(),
		Code:         awslambda.Code_FromAsset(jsii.String("base-lambda/build"), &awss3assets.AssetOptions{}),
		Handler:      jsii.String("bootstrap"),
		Environment:  props.EnvVariables,
	})
}
