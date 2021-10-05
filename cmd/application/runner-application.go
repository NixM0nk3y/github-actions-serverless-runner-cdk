package main

import (
	"fmt"
	"github-actions-serverless-runner/pkg/stacks"
	"log"
	"os"
	"strings"

	"github.com/aws/aws-cdk-go/awscdk"
	"github.com/aws/jsii-runtime-go"

	"github.com/kelseyhightower/envconfig"
	"go.uber.org/zap"
)

var logger *zap.SugaredLogger

func main() {

	log.Print("Starting Application Build")

	app := awscdk.NewApp(&awscdk.AppProps{
		AnalyticsReporting: jsii.Bool(false),
	})

	stackProps := stacks.GithubActionsServerlessRunnerProps{
		StackProps: awscdk.StackProps{
			Env: env(),
		},
	}

	err := envconfig.Process("cdk", &stackProps)

	if err != nil {
		log.Fatal(err.Error())
	}

	id := fmt.Sprintf("%s%sRunnerStack", strings.Title(stackProps.Tenant), strings.Title(stackProps.Environment))

	stacks.GithubActionsServerlessRunnerStack(app, id, &stackProps)

	app.Synth(nil)
}

// env determines the AWS environment (account+region) in which our stack is to
// be deployed. For more information see: https://docs.aws.amazon.com/cdk/latest/guide/environments.html
func env() *awscdk.Environment {
	return &awscdk.Environment{
		Account: jsii.String(os.Getenv("CDK_DEFAULT_ACCOUNT")),
		Region:  jsii.String(os.Getenv("CDK_DEFAULT_REGION")),
	}
}
