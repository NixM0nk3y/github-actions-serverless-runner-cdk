package builder

import (
	"io/ioutil"
	"log"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awscodebuild"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsecrassets"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsiam"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"gopkg.in/yaml.v3"
)

type BuilderProps struct {
	Tenant      string ``
	Environment string ``
}

type Buildspec *map[string]interface{}

func Builder(scope constructs.Construct, id string, props *BuilderProps) awscodebuild.PipelineProject {

	construct := constructs.NewConstruct(scope, &id)

	role := awsiam.NewRole(construct, jsii.String("Role"), &awsiam.RoleProps{
		AssumedBy: awsiam.NewServicePrincipal(jsii.String("codebuild.amazonaws.com"), &awsiam.ServicePrincipalOpts{}),
		Path:      jsii.String("/service-roles/builder/"),
	})

	content, err := ioutil.ReadFile("./resources/buildspec.yml") // the file is inside the local directory

	if err != nil {
		log.Fatalf("error buildspec file: %v", err)
	}

	var buildspec Buildspec
	err = yaml.Unmarshal(content, &buildspec)

	if err != nil {
		log.Fatalf("unable to unmarshal data: %v", err)
	}

	return awscodebuild.NewPipelineProject(construct, jsii.String("Project"), &awscodebuild.PipelineProjectProps{
		Description:          jsii.String("Github Actions Build Runner"),
		ConcurrentBuildLimit: jsii.Number(5),
		Role:                 role,
		Timeout:              awscdk.Duration_Minutes(jsii.Number(10)),
		BuildSpec:            awscodebuild.BuildSpec_FromObjectToYaml(buildspec),
		Environment: &awscodebuild.BuildEnvironment{
			ComputeType: awscodebuild.ComputeType_SMALL,
			Privileged:  jsii.Bool(true),
			BuildImage: awscodebuild.LinuxBuildImage_FromAsset(construct, jsii.String("BuildImage"), &awsecrassets.DockerImageAssetProps{
				Directory: jsii.String("./resources/actionrunner"),
			}),
		},
	})
}
