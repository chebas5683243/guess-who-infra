package customconstructs

import (
	"github.com/chebas5683243/guess-who-infra/config"
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

type DynamoDBTable struct {
	awsdynamodb.TableV2
	baseName *string
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

func (d *DynamoDBConstruct) CreateTable(props *DynamoDBTableProps) *DynamoDBTable {
	if props == nil {
		panic("dynamoDB table configuration missing")
	}

	if props.TableName == nil {
		panic("missing table name for dynamodb table")
	}

	tableName := config.StackName + "__" + *props.TableName + d.env

	table := awsdynamodb.NewTableV2(d, &tableName, &awsdynamodb.TablePropsV2{
		TableName:              &tableName,
		PartitionKey:           props.PartitionKey,
		SortKey:                props.SortKey,
		LocalSecondaryIndexes:  props.LocalSecondaryIndexes,
		GlobalSecondaryIndexes: props.GlobalSecondaryIndexes,
		Billing:                awsdynamodb.Billing_OnDemand(nil),
		RemovalPolicy:          awscdk.RemovalPolicy_DESTROY,
	})

	return &DynamoDBTable{
		TableV2:  table,
		baseName: props.TableName,
	}
}

func (d *DynamoDBConstruct) GrantAccessToLambda(table *DynamoDBTable, lambda awslambda.Function) {
	table.GrantReadWriteData(lambda)

	tableEnvVariable := *table.baseName + "_TABLE"
	lambda.AddEnvironment(&tableEnvVariable, table.TableName(), nil)
}
