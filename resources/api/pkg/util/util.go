package util

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"strings"

	"api/pkg/log"

	"go.uber.org/zap"
)

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
