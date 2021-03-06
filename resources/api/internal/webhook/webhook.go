package webhook

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"api/internal/config"
	"api/pkg/github"

	"api/pkg/log"
	"api/pkg/util"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	gh "github.com/google/go-github/v39/github"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/codebuild"
	"github.com/aws/aws-sdk-go-v2/service/codebuild/types"
	"github.com/aws/smithy-go"
	"github.com/go-chi/render"
	"go.uber.org/zap"
)

type ErrResponse struct {
	Err            error `json:"-"` // low-level runtime error
	HTTPStatusCode int   `json:"-"` // http response status code

	StatusText string `json:"status"`          // user-level status message
	ErrorText  string `json:"error,omitempty"` // application-level error message, for debugging
}

func (e *ErrResponse) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, e.HTTPStatusCode)
	return nil
}

func ErrInvalidRequest(err error) render.Renderer {
	return &ErrResponse{
		Err:            err,
		HTTPStatusCode: 400,
		StatusText:     "Invalid request.",
		ErrorText:      err.Error(),
	}
}

type webhookResponse struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}

func (u *webhookResponse) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

type workflowEvent struct {
	gh.WorkflowJobEvent
}

func (g *workflowEvent) Bind(r *http.Request) error {
	return nil
}

func StartBuilder(ctx context.Context, builder string, event *workflowEvent) error {
	logger := log.LoggerWithLambdaRqID(ctx)

	logger.Debug("startBuilder")

	repoData := strings.Split(*event.Repo.FullName, "/")
	owner := repoData[0]
	repo := repoData[1]

	logger = logger.With(
		zap.Reflect("owner", owner),
		zap.Reflect("repo", repo),
		zap.Reflect("hash", event.WorkflowJob.HeadSHA),
		zap.Reflect("name", event.WorkflowJob.Name),
	)

	logger.Info("minting new runner token")
	appKey := []byte(config.GetConfig(ctx).GithubAppKey)
	runnerToken := github.GenerateRunnerToken(ctx, owner, repo, config.GetConfig(ctx).GithubAppID, config.GetConfig(ctx).GithubInstallationID, appKey)

	logger.Info("starting builder")
	cfg, err := awsconfig.LoadDefaultConfig(ctx)
	if err != nil {
		logger.Error("failed to load configuration", zap.Error(err))
	}

	svc := codebuild.NewFromConfig(cfg)

	output, err := svc.StartBuild(ctx, &codebuild.StartBuildInput{
		ProjectName: aws.String(builder),
		ArtifactsOverride: &types.ProjectArtifacts{
			Type: types.ArtifactsTypeNoArtifacts,
		},
		SourceTypeOverride: types.SourceTypeNoSource,
		EnvironmentVariablesOverride: []types.EnvironmentVariable{
			{
				Name:  aws.String("GITHUB_REPO"),
				Type:  types.EnvironmentVariableTypePlaintext,
				Value: aws.String(*event.Repo.FullName),
			},
			{
				Name:  aws.String("RUNNER_TOKEN"),
				Type:  types.EnvironmentVariableTypePlaintext,
				Value: aws.String(runnerToken),
			},
		},
	})

	if err != nil {
		var oe *smithy.OperationError
		if errors.As(err, &oe) {
			logger.Fatal(fmt.Sprintf("failed to call service: %s, operation: %s, error: %v", oe.Service(), oe.Operation(), oe.Unwrap()))
		}
		return err
	}

	logger.Info("build started", zap.String("id", *output.Build.Id), zap.String("status", string(output.Build.BuildStatus)))

	return err
}

func ProcessActionEvent(ctx context.Context, event *workflowEvent) {
	logger := log.LoggerWithLambdaRqID(ctx)

	logger.Debug("processActionEvent")

	logger = logger.With(
		zap.Reflect("repo", event.Repo.FullName),
		zap.Reflect("hash", event.WorkflowJob.HeadSHA),
		zap.Reflect("name", event.WorkflowJob.Name),
	)

	// we're only interested in queued events for self hosted runners
	if *event.Action == "queued" && util.Contains(event.WorkflowJob.Labels, "self-hosted") {
		logger.Info("processing self hosted github action", zap.String("action", *event.Action))
		StartBuilder(ctx, config.GetConfig(ctx).BuilderProject, event)
	} else {
		logger.Info("ignoring github action", zap.String("action", *event.Action))
	}
}

func EventHandler(w http.ResponseWriter, r *http.Request) {

	logger := log.LoggerWithLambdaRqID(r.Context())

	logger.Debug("EventHandler")

	util.RequestDump(r)

	// does the webhook signiture match ours?
	if valid := util.IsValidSignature(r, config.GetConfig(r.Context()).GithubHookSecret); !valid {
		logger.Error("failed to validate signature")
		render.Render(w, r, ErrInvalidRequest(errors.New("failed to validate secret")))
		return
	}

	// attempt to unmarshal our event
	event := &workflowEvent{}

	if err := render.Bind(r, event); err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	if event.Action != nil {
		logger.Info("event validated, processing")
		ProcessActionEvent(r.Context(), event)
	} else {
		logger.Warn("skipping event , doesn't seem to be workflow type")
	}

	render.Render(w, r, &webhookResponse{
		Status:  200,
		Message: "OK",
	})
}
