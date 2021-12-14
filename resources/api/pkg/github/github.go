package github

import (
	"api/pkg/log"
	"context"
	"net/http"
	"time"

	"github.com/aws/aws-xray-sdk-go/xray"
	"github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/google/go-github/v39/github"
	"go.uber.org/zap"
)

type Client struct {
	Ctx       context.Context
	GitHub    *github.Client
	RepoOwner string
}

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

func InitAppClient(ctx context.Context, installationId int64, githubAppID int64, githubPrivateKey []byte) (*github.Client, error) {
	tr := http.DefaultTransport

	// xray wrap a http client
	xtr := xray.RoundTripper(tr)

	itr, err := ghinstallation.New(
		xtr,
		githubAppID,
		int64(installationId),
		githubPrivateKey,
	)
	if err != nil {
		return nil, err
	}

	c := github.NewClient(&http.Client{Transport: itr})

	return c, nil
}

func GenerateRunnerToken(ctx context.Context, owner, repo string, appID, installationID int64, appKey []byte) string {

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

	client, err := InitAppClient(ctx, installationID, appID, appKey)
	if err != nil {
		logger.Panic("github app init", zap.Error(err))
	}

	// mint a new token
	logger.Info("fetching new runner token")
	newToken, _, err := client.Actions.CreateRegistrationToken(ctx, owner, repo)
	if err != nil {
		logger.Panic("github error", zap.Error(err))
	}

	// cache our new token and return it
	runnerTokens[cacheKey] = newToken

	return newToken.GetToken()
}
