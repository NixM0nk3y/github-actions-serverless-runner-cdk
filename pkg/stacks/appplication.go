package stacks

import (
	"github-actions-serverless-runner/pkg/builder"
	"github-actions-serverless-runner/pkg/webhook"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/constructs-go/constructs/v10"
)

type GithubActionsServerlessRunnerProps struct {
	Tenant      string            `envconfig:"TENANT" default:"openenterprise"`
	Environment string            `envconfig:"ENVIRONMENT" default:"staging"`
	AuthToken   string            `envconfig:"AUTH_TOKEN" default:"AAAAAAA"`
	HookSecret  string            `envconfig:"HOOK_SECRET" default:"BBBBBBB"`
	StackProps  awscdk.StackProps ``
}

func GithubActionsServerlessRunnerStack(scope constructs.Construct, id string, props *GithubActionsServerlessRunnerProps) awscdk.Stack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}

	stack := awscdk.NewStack(scope, &id, &sprops)

	pipeline := builder.Builder(stack, "Builder", &builder.BuilderProps{
		Tenant:      props.Tenant,
		Environment: props.Environment,
	})

	webhook.Webhook(stack, "WebHook", &webhook.WebhookProps{
		Tenant:           props.Tenant,
		Environment:      props.Environment,
		Builder:          pipeline,
		GithubAuthToken:  props.AuthToken,
		GithubHookSecret: props.HookSecret,
	})

	return stack
}
