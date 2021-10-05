# Serverless Github Action Runner AWS CDK Stack 

A self hosted github action runner using codebuild as compute.

# Stack Setup

```bash
$ AUTH_TOKEN=ghp_mytoken HOOK_SECRET=rand0mgener4teds3cret make deploy
go build -v ./cmd/application
🛠️  cmd/application done
✓  stacks/build done
✓  build done
cdk deploy --app ./application
2021/09/29 20:45:58 Starting Application Build
Bundling asset OpenenterpriseDevelopmentRunnerStack/WebHook/Lambda/Code/Stage...
OpenenterpriseDevelopmentRunnerStack: deploying...
[0%] start: Publishing 5f5379576f3bed77676254e6a014479f428d0584dc49711933e0b9f7f6d43455:074705540277-eu-west-1
[25%] success: Published 5f5379576f3bed77676254e6a014479f428d0584dc49711933e0b9f7f6d43455:074705540277-eu-west-1
[25%] start: Publishing f3b4457410e8875dce33608b6046e993c1ba6def4d00e9de1ed9681517a35e45:074705540277-eu-west-1
[50%] success: Published f3b4457410e8875dce33608b6046e993c1ba6def4d00e9de1ed9681517a35e45:074705540277-eu-west-1
[50%] start: Publishing 7fab770ff105ad542336579d216faf83252c1d2092ec9e084cf45c1d90f1dca8:074705540277-eu-west-1
[75%] success: Published 7fab770ff105ad542336579d216faf83252c1d2092ec9e084cf45c1d90f1dca8:074705540277-eu-west-1
[75%] start: Publishing 2dac7654f9d8a638fd7269ff35d4214c1a53abe99ba7a1f4026bb566dc54021c:074705540277-eu-west-1
[100%] success: Published 2dac7654f9d8a638fd7269ff35d4214c1a53abe99ba7a1f4026bb566dc54021c:074705540277-eu-west-1
OpenenterpriseDevelopmentRunnerStack: creating CloudFormation changeset...

 ✅  OpenenterpriseDevelopmentRunnerStack

Stack ARN:
arn:aws:cloudformation:eu-west-1:074705540277:stack/OpenenterpriseDevelopmentRunnerStack/a17b7750-1bd5-11ec-95ee-0a661266e345
🛠️  deploy/application done
✓  deploy done
```

The once the stack is deployed setup the github repo to recieve "Workflow jobs" events using the generated API GW URL and secret

# License

MIT

## Useful commands

 * `make deploy`          deploy this stack to your default AWS account/region
 * `make security/scan`   run a security scan
 * `make test`            run unit tests
