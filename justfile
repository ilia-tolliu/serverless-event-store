default:
    @just --list

clean:
    rm -rf ./build && rm -f function.zip

build: clean
    GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="-s -w" -tags lambda.norpc -o ./build/bootstrap ./cmd/event_store_lambda
    chmod 644 ./build/bootstrap
    cp -r swagger_ui ./build
    cp openapi_spec.json ./build
    (cd ./build && zip -r ../function.zip .)

app_mode := "staging"

deploy: build
    (cd _infrastructure/aws-event-store && ES_APP_MODE={{app_mode}} cdk deploy)

describe:
    aws cloudformation describe-stacks --stack-name AwsEventStoreStack-{{app_mode}} | jq '.Stacks | .[] | .Outputs | reduce .[] as $i ({}; .[$i.OutputKey] = $i.OutputValue)'

diff:
    (cd _infrastructure/aws-event-store && ES_APP_MODE={{app_mode}} cdk diff)

test:
    EVENT_STORE_MODE={{app_mode}} go test ./test -count=1

# Requires godoc tool to be installed:
# $ go install golang.org/x/tools/cmd/godoc@latest
#
# Build godoc documentation and make it awailable at http://localhost:8000
doc:
    godoc -http=:8000