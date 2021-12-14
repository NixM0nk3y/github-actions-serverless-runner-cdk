package stacks

import (
	"github-actions-serverless-runner/pkg/builder"
	"github-actions-serverless-runner/pkg/webhook"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/constructs-go/constructs/v10"
)

type GithubActionsServerlessRunnerProps struct {
	Tenant               string            `envconfig:"TENANT" default:"openenterprise"`
	Environment          string            `envconfig:"ENVIRONMENT" default:"staging"`
	GithubAppID          string            `envconfig:"GITHUB_APP_ID" required:"true"`
	GithubInstallationID string            `envconfig:"GITHUB_INSTALLATION_ID" required:"true"`
	GithubAppKeyPath     string            `envconfig:"GITHUB_APP_KEY_PATH" required:"true"`
	GithubHookSecret     string            `envconfig:"GITHUB_HOOK_SECRET" default:"BBBBBBB"`
	StackProps           awscdk.StackProps ``
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
		Tenant:               props.Tenant,
		Environment:          props.Environment,
		Builder:              pipeline,
		GithubAppID:          props.GithubAppID,
		GithubInstallationID: props.GithubInstallationID,
		GithubAppKeyPath:     props.GithubAppKeyPath,
		GithubHookSecret:     props.GithubHookSecret,
	})

	return stack
}
