package webhook

import (
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-cdk-go/awscdk"
	"github.com/aws/aws-cdk-go/awscdk/awsapigatewayv2"
	"github.com/aws/aws-cdk-go/awscdk/awsapigatewayv2integrations"
	"github.com/aws/aws-cdk-go/awscdk/awscodebuild"
	"github.com/aws/aws-cdk-go/awscdk/awsiam"
	"github.com/aws/aws-cdk-go/awscdk/awslambda"
	"github.com/aws/aws-cdk-go/awscdk/awslambdago"
	"github.com/aws/aws-cdk-go/awscdk/awslogs"
	"github.com/aws/constructs-go/constructs/v3"
	"github.com/aws/jsii-runtime-go"
)

type WebhookProps struct {
	Tenant           string                       ``
	Environment      string                       ``
	GithubAuthToken  string                       ``
	GithubHookSecret string                       ``
	Builder          awscodebuild.PipelineProject ``
}

func Webhook(scope constructs.Construct, id string, props *WebhookProps) awscdk.Construct {

	construct := awscdk.NewConstruct(scope, &id)

	buildNumber, ok := os.LookupEnv("CODEBUILD_BUILD_NUMBER")
	if !ok {
		// default version
		buildNumber = "0"
	}

	sourceVersion, ok := os.LookupEnv("CODEBUILD_RESOLVED_SOURCE_VERSION")
	if !ok {
		sourceVersion = "unknown"
	}

	buildDate, ok := os.LookupEnv("BUILD_DATE")
	if !ok {
		t := time.Now()
		buildDate = t.Format("20060102")
	}

	// Go build options
	bundlingOptions := &awslambdago.BundlingOptions{
		GoBuildFlags: &[]*string{jsii.String(fmt.Sprintf(`-ldflags "-s -w
			-X api/pkg/version.Version=1.0.%s
			-X api/pkg/version.BuildHash=%s
			-X api/pkg/version.BuildDate=%s
			"`,
			buildNumber,
			sourceVersion,
			buildDate,
		)),
		},
	}

	// webhook lambda
	webHookLambda := awslambdago.NewGoFunction(construct, jsii.String("Lambda"), &awslambdago.GoFunctionProps{
		Runtime:      awslambda.Runtime_GO_1_X(),
		Entry:        jsii.String("resources/api/cmd/webhook"),
		Bundling:     bundlingOptions,
		Tracing:      awslambda.Tracing_ACTIVE,
		LogRetention: awslogs.RetentionDays_ONE_WEEK,
		//Architectures: [awslambda.Architecture_ARM_64()],
		Environment: &map[string]*string{
			"HOOK_SECRET":     jsii.String(props.GithubHookSecret),
			"AUTH_TOKEN":      jsii.String(props.GithubAuthToken),
			"LOG_LEVEL":       jsii.String("DEBUG"),
			"BUILDER_PROJECT": props.Builder.ProjectName(),
		},
		ModuleDir: jsii.String("resources/api/go.mod"),
	})

	webHookLambda.Role().AddToPrincipalPolicy(awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
		Sid:     jsii.String("PermitLambdaStartBuild"),
		Effect:  awsiam.Effect_ALLOW,
		Actions: jsii.Strings("CodeBuild:StartBuild"),
		Resources: &[]*string{
			props.Builder.ProjectArn(),
		},
	}))

	//
	httpapi := awsapigatewayv2.NewHttpApi(construct, jsii.String("WebHookAPI"), &awsapigatewayv2.HttpApiProps{})

	// POST
	webhookPostIntegration := awsapigatewayv2integrations.NewLambdaProxyIntegration(&awsapigatewayv2integrations.LambdaProxyIntegrationProps{
		Handler:              webHookLambda,
		PayloadFormatVersion: awsapigatewayv2.PayloadFormatVersion_VERSION_1_0(),
	})

	httpapi.AddRoutes(&awsapigatewayv2.AddRoutesOptions{
		Integration: webhookPostIntegration,
		Path:        jsii.String("/webhook/version"),
		Methods: &[]awsapigatewayv2.HttpMethod{
			awsapigatewayv2.HttpMethod_GET,
		},
	})

	httpapi.AddRoutes(&awsapigatewayv2.AddRoutesOptions{
		Integration: webhookPostIntegration,
		Path:        jsii.String("/webhook/event"),
		Methods: &[]awsapigatewayv2.HttpMethod{
			awsapigatewayv2.HttpMethod_POST,
		},
	})

	return construct
}
