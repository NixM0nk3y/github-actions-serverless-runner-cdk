package webhook

import (
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-cdk-go/awscdk/v2/awscodebuild"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsiam"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambda"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslogs"
	"github.com/aws/aws-cdk-go/awscdkapigatewayv2alpha/v2"
	"github.com/aws/aws-cdk-go/awscdkapigatewayv2integrationsalpha/v2"
	"github.com/aws/aws-cdk-go/awscdklambdagoalpha/v2"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

type WebhookProps struct {
	Tenant           string                       ``
	Environment      string                       ``
	GithubAuthToken  string                       ``
	GithubHookSecret string                       ``
	Builder          awscodebuild.PipelineProject ``
}

func Webhook(scope constructs.Construct, id string, props *WebhookProps) constructs.Construct {

	construct := constructs.NewConstruct(scope, &id)

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
	bundlingOptions := &awscdklambdagoalpha.BundlingOptions{
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
		Environment: &map[string]*string{
			"GOARCH":      jsii.String("arm64"),
			"GO111MODULE": jsii.String("on"),
			"GOOS":        jsii.String("linux"),
		},
	}

	// webhook lambda
	webHookLambda := awscdklambdagoalpha.NewGoFunction(construct, jsii.String("Lambda"), &awscdklambdagoalpha.GoFunctionProps{
		Runtime:      awslambda.Runtime_PROVIDED_AL2(),
		Entry:        jsii.String("resources/api/cmd/webhook"),
		Bundling:     bundlingOptions,
		Tracing:      awslambda.Tracing_ACTIVE,
		LogRetention: awslogs.RetentionDays_ONE_WEEK,
		Architecture: awslambda.Architecture_ARM_64(),
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
	httpapi := awscdkapigatewayv2alpha.NewHttpApi(construct, jsii.String("WebHookAPI"), &awscdkapigatewayv2alpha.HttpApiProps{})

	// POST
	webhookPostIntegration := awscdkapigatewayv2integrationsalpha.NewHttpLambdaIntegration(jsii.String("WebHookInt"), webHookLambda, &awscdkapigatewayv2integrationsalpha.HttpLambdaIntegrationProps{
		PayloadFormatVersion: awscdkapigatewayv2alpha.PayloadFormatVersion_VERSION_1_0(),
	})

	httpapi.AddRoutes(&awscdkapigatewayv2alpha.AddRoutesOptions{
		Integration: webhookPostIntegration,
		Path:        jsii.String("/webhook/version"),
		Methods: &[]awscdkapigatewayv2alpha.HttpMethod{
			awscdkapigatewayv2alpha.HttpMethod_GET,
		},
	})

	httpapi.AddRoutes(&awscdkapigatewayv2alpha.AddRoutesOptions{
		Integration: webhookPostIntegration,
		Path:        jsii.String("/webhook/event"),
		Methods: &[]awscdkapigatewayv2alpha.HttpMethod{
			awscdkapigatewayv2alpha.HttpMethod_POST,
		},
	})

	return construct
}
