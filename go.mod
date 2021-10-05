module github-actions-serverless-runner

go 1.16

require (
	github.com/aws/aws-cdk-go/awscdk v1.125.0-devpreview
	github.com/aws/constructs-go/constructs/v3 v3.3.99
	github.com/aws/jsii-runtime-go v1.34.0
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/onsi/ginkgo v1.16.4 // indirect
	github.com/onsi/gomega v1.14.0 // indirect
	github.com/stretchr/testify v1.7.0

	// for testing
	github.com/tidwall/gjson v1.7.4
	go.uber.org/zap v1.18.1
	gopkg.in/go-playground/assert.v1 v1.2.1
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
)
