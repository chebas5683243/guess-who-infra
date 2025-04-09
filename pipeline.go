package main

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awscodebuild"
	"github.com/aws/aws-cdk-go/awscdk/v2/awscodepipeline"
	"github.com/aws/aws-cdk-go/awscdk/v2/awscodepipelineactions"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

type PipelineStackProps struct {
	awscdk.StackProps
	GitHubOwner string
	GitHubRepo  string
}

func NewPipelineStack(scope constructs.Construct, id string, props *PipelineStackProps) awscdk.Stack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	pipeline := awscodepipeline.NewPipeline(stack, jsii.String(AppName+"Pipeline"), &awscodepipeline.PipelineProps{
		PipelineName: jsii.String(AppName + "Pipeline"),
	})

	sourceOutput := awscodepipeline.NewArtifact(jsii.String("SourceArtifact"), nil)
	sourceAction := awscodepipelineactions.NewGitHubSourceAction(&awscodepipelineactions.GitHubSourceActionProps{
		ActionName: jsii.String("GitHub_Source"),
		Owner:      jsii.String(props.GitHubOwner),
		Repo:       jsii.String(props.GitHubRepo),
		Branch:     jsii.String("main"),
		Output:     sourceOutput,
	})

	pipeline.AddStage(&awscodepipeline.StageOptions{
		StageName: jsii.String("Source"),
		Actions:   &[]awscodepipeline.IAction{sourceAction},
	})

	buildOutput := awscodepipeline.NewArtifact(jsii.String("BuildArtifact"), nil)
	buildProject := awscodebuild.NewPipelineProject(stack, jsii.String(AppName+"Build"), &awscodebuild.PipelineProjectProps{
		ProjectName: jsii.String(AppName + "Build"),
		Environment: &awscodebuild.BuildEnvironment{
			BuildImage: awscodebuild.LinuxBuildImage_STANDARD_5_0(),
		},
		BuildSpec: awscodebuild.BuildSpec_FromObject(&map[string]any{
			"version": "0.2",
			"phases": map[string]any{
				"build": map[string]any{
					"commands": []string{
						"cd lambda",
						"go test ./...",
						"go build -o main",
					},
				},
			},
			"artifacts": map[string]any{
				"files": []string{
					"lambda/main",
				},
			},
		}),
	})

	buildAction := awscodepipelineactions.NewCodeBuildAction(&awscodepipelineactions.CodeBuildActionProps{
		ActionName: jsii.String("Build4All"),
		Project:    buildProject,
		Input:      sourceOutput,
		Outputs:    &[]awscodepipeline.Artifact{buildOutput},
	})

	pipeline.AddStage(&awscodepipeline.StageOptions{
		StageName: jsii.String("Build"),
		Actions:   &[]awscodepipeline.IAction{buildAction},
	})

	devDeployAction := awscodepipelineactions.NewCloudFormationCreateUpdateStackAction(&awscodepipelineactions.CloudFormationCreateUpdateStackActionProps{
		ActionName:       jsii.String("DeployToLambda"),
		TemplatePath:     buildOutput.AtPath(jsii.String("template.yaml")),
		StackName:        jsii.String(AppName + "InfraStackDev"),
		AdminPermissions: jsii.Bool(true),
	})

	pipeline.AddStage(&awscodepipeline.StageOptions{
		StageName: jsii.String("Development"),
		Actions:   &[]awscodepipeline.IAction{devDeployAction},
	})

	approvalAction := awscodepipelineactions.NewManualApprovalAction(&awscodepipelineactions.ManualApprovalActionProps{
		ActionName: jsii.String("ProductionApproval"),
	})

	prodDeployAction := awscodepipelineactions.NewCloudFormationCreateUpdateStackAction(&awscodepipelineactions.CloudFormationCreateUpdateStackActionProps{
		ActionName:       jsii.String("DeployToLambda"),
		TemplatePath:     buildOutput.AtPath(jsii.String("template.yaml")),
		StackName:        jsii.String(AppName + "InfraStackProd"),
		AdminPermissions: jsii.Bool(true),
	})

	pipeline.AddStage(&awscodepipeline.StageOptions{
		StageName: jsii.String("Production"),
		Actions:   &[]awscodepipeline.IAction{approvalAction, prodDeployAction},
	})

	return stack
}
