package main

import (
	"fmt"
	"os"
	"path"
	"runtime"

	"github-actions-serverless-runner/pkg/stacks"
	"github-actions-serverless-runner/security/utils"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/jsii-runtime-go"
	"github.com/kelseyhightower/envconfig"
)

// run test under the pkg root directory
func init() {
	_, filename, _, _ := runtime.Caller(0)
	dir := path.Join(path.Dir(filename), "..")
	err := os.Chdir(dir)
	if err != nil {
		panic(err)
	}
}

func main() {

	// generate our test harness
	app, dir := utils.SetupStack()

	// clean up result
	defer os.RemoveAll(dir)

	stackProps := stacks.GithubActionsServerlessRunnerProps{
		StackProps: awscdk.StackProps{
			Env: &awscdk.Environment{
				Account: jsii.String("123456789012"),
				Region:  jsii.String("eu-west-1"),
			},
		}}

	envconfig.Process("cdk", &stackProps)

	stacks.GithubActionsServerlessRunnerStack(app, "ActionStack", &stackProps)

	// run our scan
	results := utils.SecurityScan(app)

	returnCode := 0

	for _, v := range results {

		if v.Result != 0 {
			returnCode = v.Result
		}

		fmt.Println(v.Report)
	}

	os.Exit(returnCode)
}
