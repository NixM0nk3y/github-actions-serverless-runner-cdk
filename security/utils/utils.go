package utils

import (
	"fmt"
	"io/ioutil"
	"log"
	"os/exec"
	"path"
	"strings"
	"syscall"

	"github.com/aws/aws-cdk-go/awscdk"
	"github.com/aws/jsii-runtime-go"
)

type SecurityScanResult struct {
	Result int
	Report string
}

func SecurityScan(app awscdk.App) (results []*SecurityScanResult) {
	// synth our app
	app.Synth(nil)

	files, err := ioutil.ReadDir(*app.AssetOutdir())
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".template.json") {
			result := runGuard(path.Join(*app.AssetOutdir(), file.Name()))
			results = append(results, result)
		}
	}

	return
}

func runGuard(template string) *SecurityScanResult {

	exitCode := 0

	cmd := exec.Command("cfn-guard", "validate", "--show-summary", "none", "--output-format", "yaml", "--data", template, "--rules", "./security/rules")

	out, err := cmd.CombinedOutput()

	if err != nil {
		exitCode = processExitCode(err)
	}

	return &SecurityScanResult{
		Result: exitCode,
		Report: string(out),
	}
}

func SetupStack() (awscdk.App, string) {
	// generate temporary directory
	dir, err := ioutil.TempDir("", "cdk")
	if err != nil {
		log.Fatal(err)
	}

	// generate our parent stack
	app := awscdk.NewApp(&awscdk.AppProps{
		AnalyticsReporting: jsii.Bool(false),
		Outdir:             jsii.String(dir),
	})

	return app, dir
}

func getExitCode(err error) (int, error) {
	exitCode := 0
	if exiterr, ok := err.(*exec.ExitError); ok {
		if procExit := exiterr.Sys().(syscall.WaitStatus); ok {
			return procExit.ExitStatus(), nil
		}
	}
	return exitCode, fmt.Errorf("failed to get exit code")
}

func processExitCode(err error) (exitCode int) {
	if err != nil {
		var exiterr error
		if exitCode, exiterr = getExitCode(err); exiterr != nil {
			// TODO: Fix this so we check the error's text.
			// we've failed to retrieve exit code, so we set it to 127
			exitCode = 127
		}
	}
	return
}
