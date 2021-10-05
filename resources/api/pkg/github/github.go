package github

import (
	"api/pkg/log"
	"context"
	"time"

	"github.com/google/go-github/v39/github"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
)

type CacheKey struct {
	owner string
	repo  string
}

type tokenCache map[CacheKey]*github.RegistrationToken

var runnerTokens tokenCache

func init() {
	// initialise our cache
	runnerTokens = make(tokenCache)
}

func GenerateRunnerToken(ctx context.Context, owner, repo, token string) string {

	logger := log.LoggerWithLambdaRqID(ctx).With(
		zap.Reflect("repo", repo),
		zap.Reflect("owner", owner),
	)

	logger.Debug("GenerateRunnerToken")

	// cache tokens between runs
	cacheKey := CacheKey{
		owner: owner,
		repo:  repo,
	}

	// have we got a valid token cached (expire every 60 mins) - return that
	if token, ok := runnerTokens[cacheKey]; ok {
		if time.Until(token.ExpiresAt.Time).Minutes() > 2.0 {
			logger.Info("returning cached runner token", zap.Float64("expires", time.Until(token.ExpiresAt.Time).Minutes()))
			return token.GetToken()
		}
	}

	// create our o-authed client
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	// mint a new token
	logger.Info("fetching new runner token")
	newToken, _, err := client.Actions.CreateRegistrationToken(ctx, owner, repo)
	if err != nil {
		logger.Error("github error", zap.Error(err))
	}

	// cache our new token and return it
	runnerTokens[cacheKey] = newToken

	return newToken.GetToken()
}
