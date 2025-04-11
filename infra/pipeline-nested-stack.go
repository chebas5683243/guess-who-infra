package infra

import (
	"github.com/chebas5683243/guess-who-infra/config"
	"github.com/chebas5683243/guess-who-infra/environment"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awscodebuild"
	"github.com/aws/aws-cdk-go/awscdk/v2/awscodepipeline"
	"github.com/aws/aws-cdk-go/awscdk/v2/awscodepipelineactions"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsiam"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

type GithubConfig struct {
	Owner  string
	Repo   string
	Branch string
	Token  string
}

type Stacks struct {
	Development *EnvironmentNestedStack
	Production  *EnvironmentNestedStack
}

type PipelineNestedStackProps struct {
	awscdk.NestedStackProps
	GithubConfig
	Stacks
}

type PipelineNestedStack struct {
	awscdk.NestedStack
	githubConfig GithubConfig
	stacks       Stacks
	pipeline     awscodepipeline.Pipeline
	sourceOutput awscodepipeline.Artifact
	buildOutput  awscodepipeline.Artifact
}

func NewPipelineNestedStack(scope constructs.Construct, id string, props *PipelineNestedStackProps) *PipelineNestedStack {
	nestedStack := &PipelineNestedStack{}

	if props == nil {
		panic("environment nested stack props missing")
	}

	awscdk.NewNestedStack_Override(nestedStack, scope, jsii.String("PipelineNestedStack"), &props.NestedStackProps)
	nestedStack.stacks = props.Stacks
	nestedStack.githubConfig = props.GithubConfig

	return nestedStack
}

func (nestedStack *PipelineNestedStack) CreateInfraestructure() {
	nestedStack.createPipeline()
	nestedStack.createSourceStage()
	nestedStack.createBuildStage()
	nestedStack.createDevelopmentStage()
	nestedStack.createProductionStage()
}

func (nestedStack *PipelineNestedStack) createPipeline() {
	pipelineName := nestedStack.generateStackResourceName("Pipeline")
	nestedStack.pipeline = awscodepipeline.NewPipeline(nestedStack, &pipelineName, &awscodepipeline.PipelineProps{
		PipelineName: &pipelineName,
	})
}

func (nestedStack *PipelineNestedStack) createSourceStage() {
	artifactName := nestedStack.generateStackResourceName("SourceArtifact")
	sourceOutput := awscodepipeline.NewArtifact(&artifactName, nil)

	sourceAction := awscodepipelineactions.NewGitHubSourceAction(&awscodepipelineactions.GitHubSourceActionProps{
		ActionName: jsii.String("Source"),
		Owner:      &nestedStack.githubConfig.Owner,
		Repo:       &nestedStack.githubConfig.Repo,
		Branch:     &nestedStack.githubConfig.Branch,
		OauthToken: awscdk.SecretValue_UnsafePlainText(&nestedStack.githubConfig.Token),
		Output:     sourceOutput,
	})

	nestedStack.pipeline.AddStage(&awscodepipeline.StageOptions{
		StageName: jsii.String("Source"),
		Actions:   &[]awscodepipeline.IAction{sourceAction},
	})
	nestedStack.sourceOutput = sourceOutput
}

func (nestedStack *PipelineNestedStack) createBuildStage() {
	artifactName := nestedStack.generateStackResourceName("BuildArtifact")
	buildOutput := awscodepipeline.NewArtifact(&artifactName, nil)

	buildProjectName := nestedStack.generateStackResourceName("BuildProject")
	buildProject := awscodebuild.NewPipelineProject(nestedStack, &buildProjectName,
		&awscodebuild.PipelineProjectProps{
			ProjectName: &buildProjectName,
			Environment: &awscodebuild.BuildEnvironment{
				BuildImage: awscodebuild.LinuxBuildImage_STANDARD_5_0(),
			},
			BuildSpec: awscodebuild.BuildSpec_FromObject(&map[string]any{
				"version": "0.2",
				"phases": map[string]any{
					"pre_build": map[string]any{
						"commands": []string{
							"go mod tidy",
						},
					},
					"build": map[string]any{
						"commands": []string{
							"go test ./...",
							"GOOS=linux GOARCH=amd64 go build -o build/bootstrap",
							"chmod +x build/bootstrap",
						},
					},
				},
				"artifacts": map[string]any{
					"files": []string{
						"build/bootstrap",
					},
				},
			}),
		})

	buildAction := awscodepipelineactions.NewCodeBuildAction(&awscodepipelineactions.CodeBuildActionProps{
		ActionName: jsii.String("Build4All"),
		Project:    buildProject,
		Input:      nestedStack.sourceOutput,
		Outputs:    &[]awscodepipeline.Artifact{buildOutput},
	})

	nestedStack.pipeline.AddStage(&awscodepipeline.StageOptions{
		StageName: jsii.String("Build"),
		Actions:   &[]awscodepipeline.IAction{buildAction},
	})
	nestedStack.buildOutput = buildOutput
}

func (nestedStack *PipelineNestedStack) createDevelopmentStage() {
	devDeployAction := nestedStack.createDeploymentAction(environment.Development, nil)

	nestedStack.pipeline.AddStage(&awscodepipeline.StageOptions{
		StageName: jsii.String("Development"),
		Actions:   &[]awscodepipeline.IAction{devDeployAction},
	})
}

func (nestedStack *PipelineNestedStack) createProductionStage() {
	prodDeployAction := nestedStack.createDeploymentAction(environment.Production, jsii.Number(2))
	approvalAction := awscodepipelineactions.NewManualApprovalAction(
		&awscodepipelineactions.ManualApprovalActionProps{
			ActionName: jsii.String("ProductionApproval"),
			RunOrder:   jsii.Number(1),
		})

	nestedStack.pipeline.AddStage(&awscodepipeline.StageOptions{
		StageName: jsii.String("Production"),
		Actions:   &[]awscodepipeline.IAction{approvalAction, prodDeployAction},
	})
}

func (nestedStack *PipelineNestedStack) createDeploymentAction(env environment.Environment, runOrder *float64) awscodepipelineactions.CodeBuildAction {
	deployProjectName := nestedStack.generateStackResourceName(string(env) + "DeployProject")
	environmentStack := nestedStack.getEnvironmentStackFromEnv(env)

	deployProject := awscodebuild.NewPipelineProject(nestedStack, &deployProjectName,
		&awscodebuild.PipelineProjectProps{
			ProjectName: &deployProjectName,
			Environment: &awscodebuild.BuildEnvironment{
				BuildImage: awscodebuild.LinuxBuildImage_STANDARD_5_0(),
			},
			BuildSpec: awscodebuild.BuildSpec_FromObject(&map[string]any{
				"version": "0.2",
				"phases": map[string]any{
					"build": map[string]any{
						"commands": []string{
							"aws lambda update-function-code " +
								"--function-name " + *environmentStack.Lambda.FunctionName() +
								"--s3-bucket " + *nestedStack.buildOutput.BucketName() +
								"--s3-key " + "build/bootstrap",
						},
					},
				},
			}),
		},
	)

	deployProject.AddToRolePolicy(
		awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
			Actions:   &[]*string{jsii.String("lambda:UpdateFunctionCode")},
			Effect:    awsiam.Effect_ALLOW,
			Resources: &[]*string{environmentStack.Lambda.FunctionArn()},
		}),
	)

	return awscodepipelineactions.NewCodeBuildAction(&awscodepipelineactions.CodeBuildActionProps{
		ActionName: jsii.String("DeployToLambda"),
		Project:    deployProject,
		Input:      nestedStack.buildOutput,
		RunOrder:   runOrder,
	})
}

func (nestedStack *PipelineNestedStack) getEnvironmentStackFromEnv(env environment.Environment) *EnvironmentNestedStack {
	var environmentStack *EnvironmentNestedStack

	switch env {
	case environment.Development:
		environmentStack = nestedStack.stacks.Development
	case environment.Production:
		environmentStack = nestedStack.stacks.Production
	}

	return environmentStack
}

func (nestedStack *PipelineNestedStack) generateStackResourceName(name string) string {
	return config.StackName + name
}
