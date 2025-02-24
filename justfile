default:
    @just --list

build:
    rm -rf ./build && rm -rf function.zip
    GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="-s -w" -tags lambda.norpc -o ./build/bootstrap ./cmd/event_store_lambda
    chmod 644 ./build/bootstrap
    cp -r swagger_ui ./build
    cp openapi_spec.json ./build
    (cd ./build && zip -r ../function.zip .)

deploy: build
    (cd infrastructure/aws-event-store && cdk deploy)

describe:
    aws cloudformation describe-stacks --stack-name AwsEventStoreStack | jq '.Stacks | .[] | .Outputs | reduce .[] as $i ({}; .[$i.OutputKey] = $i.OutputValue)'
