package util

import (
	"context"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"os"
	"strings"

	"api/pkg/log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/aws/session"
	"go.uber.org/zap"
)

// GetSession is
func GetSession(ctx context.Context) (newsess client.ConfigProvider) {

	logger := log.Logger(ctx)

	config := &aws.Config{
		Logger:   &log.AWSLogger{},
		LogLevel: log.AWSLevel(),
	}

	// override endpoint supplied
	if awsEndpoint := os.Getenv("AWS_ENDPOINT"); awsEndpoint != "" {
		logger.Info(fmt.Sprintf("setting endpoint to %s", awsEndpoint))
		config.Endpoint = aws.String(awsEndpoint)
	}

	// override endpoint supplied
	if awsS3pathstyle := os.Getenv("AWS_S3_FORCEPATHSTYLE"); awsS3pathstyle != "" {
		logger.Info("setting S3 to pathstyle")
		config.S3ForcePathStyle = aws.Bool(true)
	}

	// set a default region
	if awsRegion := os.Getenv("AWS_REGION"); awsRegion == "" {
		logger.Info("setting default region")
		config.Region = aws.String("eu-west-1")
	}

	newsess, err := session.NewSession(config)

	if err != nil {
		logger.Panic("unable generate new session", zap.Error(err))
	}

	return
}

func IsValidSignature(r *http.Request, key string) bool {

	logger := log.LoggerWithLambdaRqID(r.Context())

	// Assuming a non-empty header
	gotHash := strings.SplitN(r.Header.Get("X-Hub-Signature"), "=", 2)
	if gotHash[0] != "sha1" {
		return false
	}

	body, err := r.GetBody()

	if err != nil {
		logger.Error("Cannot get the body reader", zap.Error(err))
		return false
	}

	b, err := ioutil.ReadAll(body)
	if err != nil {
		logger.Error("Cannot read the request body", zap.Error(err))
		return false
	}

	hash := hmac.New(sha1.New, []byte(key))
	if _, err := hash.Write(b); err != nil {
		logger.Error("Cannot compute the HMAC for request", zap.Error(err))
		return false
	}

	expectedHash := hex.EncodeToString(hash.Sum(nil))

	logger.Debug(fmt.Sprintf("recieved hash: %s", gotHash[1]))
	logger.Debug(fmt.Sprintf("expected hash: %s", expectedHash))

	return gotHash[1] == expectedHash
}

// Contains tells whether a contains x.
func Contains(a []string, x string) bool {
	for _, n := range a {
		if x == n {
			return true
		}
	}
	return false
}

// RequestDump is
func RequestDump(r *http.Request) {

	logger := log.LoggerWithLambdaRqID(r.Context())

	// Save a copy of this request for debugging.
	requestDump, err := httputil.DumpRequest(r, true)
	if err != nil {
		logger.Error("unable to dump request", zap.Error(err))
	}

	logger.Debug(string(requestDump))

	return
}
