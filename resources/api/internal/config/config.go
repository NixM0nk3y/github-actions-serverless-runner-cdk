package config

import (
	"api/pkg/log"
	"context"

	"github.com/kelseyhightower/envconfig"
	"go.uber.org/zap"
)

type WebhookConfig struct {
	LogLevel       string `envconfig:"LOG_LEVEL" default:"INFO"`
	AuthToken      string `envconfig:"AUTH_TOKEN" required:"true"`
	HookSecret     string `envconfig:"HOOK_SECRET" required:"true"`
	BuilderProject string `envconfig:"BUILDER_PROJECT" required:"true"`
}

// our config
var webhookConfig WebhookConfig

// The key type is unexported to prevent collisions with context keys defined in
// other packages.
type contextKey string

func (c contextKey) String() string {
	return "context key " + string(c)
}

//
var (
	contextKeyConfig = contextKey("config")
)

// ReadEnvConfig is
func ReadEnvConfig(ctx context.Context, namespace string) context.Context {

	logger := log.Logger(ctx)

	err := envconfig.Process(namespace, &webhookConfig)

	if err != nil {
		logger.Panic("unable to process environment", zap.Error(err))
	}

	return context.WithValue(ctx, contextKeyConfig, &webhookConfig)
}

// GetConfig is
func GetConfig(ctx context.Context) *WebhookConfig {

	logger := log.Logger(ctx)

	econfig, ok := ctx.Value(contextKeyConfig).(*WebhookConfig)

	if !ok {
		logger.Panic("unable to retrieve config value")
	}

	return econfig
}
