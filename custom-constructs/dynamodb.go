package customconstructs

import (
	"github.com/chebas5683243/guess-who-infra/environment"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsdynamodb"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambda"
	"github.com/aws/constructs-go/constructs/v10"
)

type DynamoDbConstructProps struct {
	environment.Environment
}

type DynamoDBConstruct struct {
	constructs.Construct
	env string
}

type DynamoDBTableProps struct {
	TableName              *string
	PartitionKey           *awsdynamodb.Attribute
	SortKey                *awsdynamodb.Attribute
	LocalSecondaryIndexes  *[]*awsdynamodb.LocalSecondaryIndexProps
	GlobalSecondaryIndexes *[]*awsdynamodb.GlobalSecondaryIndexPropsV2
}

func NewDynamoDBConstruct(scope constructs.Construct, id string, props *DynamoDbConstructProps) *DynamoDBConstruct {
	construct := &DynamoDBConstruct{}
	constructs.NewConstruct_Override(construct, scope, &id)

	if props == nil {
		panic("dynamodb construct props missing")
	}

	construct.env = string(props.Environment)
	return construct
}

func (d *DynamoDBConstruct) CreateTable(props *DynamoDBTableProps) awsdynamodb.TableV2 {
	if props == nil {
		panic("dynamoDB table configuration missing")
	}

	if props.TableName == nil {
		panic("missing table name for dynamodb table")
	}

	stack := awscdk.Stack_Of(d)
	tableName := *stack.StackName() + "__" + *props.TableName + d.env

	return awsdynamodb.NewTableV2(d, &tableName, &awsdynamodb.TablePropsV2{
		TableName:              &tableName,
		PartitionKey:           props.PartitionKey,
		SortKey:                props.SortKey,
		LocalSecondaryIndexes:  props.LocalSecondaryIndexes,
		GlobalSecondaryIndexes: props.GlobalSecondaryIndexes,
		Billing:                awsdynamodb.Billing_OnDemand(nil),
	})
}

func (d *DynamoDBConstruct) GrantAccessToLambda(table awsdynamodb.TableV2, lambda awslambda.Function) {
	table.GrantReadWriteData(lambda)

	tableEnvVariable := *table.TableName() + "_TABLE"
	lambda.AddEnvironment(&tableEnvVariable, table.TableName(), nil)
}
