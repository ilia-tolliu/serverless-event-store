package config

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/aws/aws-sdk-go-v2/service/ssm/types"
	"go.uber.org/zap"
	"strings"
)

type EsConfig struct {
	Port      string
	TableName string
}

func FromAws(ctx context.Context, mode AppMode, awsConfig aws.Config, log *zap.SugaredLogger) (*EsConfig, error) {
	path := configPrefix(mode)
	log.Infow("loading configuration", "path", path)

	params, err := loadSsmParams(ctx, awsConfig, path)
	if err != nil {
		return nil, err
	}

	port, err := extractParameter(params, "PORT")
	if err != nil {
		return nil, err
	}

	tableName, err := extractParameter(params, "DYNAMODB_TABLE_NAME")
	if err != nil {
		return nil, err
	}

	return &EsConfig{
		Port:      port,
		TableName: tableName,
	}, nil
}

func configPrefix(mode AppMode) string {
	var prefix string

	switch mode {
	case Development:
		prefix = "/development/event-store"
	case Staging:
		prefix = "/staging/event-store"
	case Production:
		prefix = "/production/event-store"
	}

	return prefix
}

func loadSsmParams(ctx context.Context, awsConfig aws.Config, path string) ([]types.Parameter, error) {
	ssmClient := ssm.NewFromConfig(awsConfig)
	ssmOutput, err := ssmClient.GetParametersByPath(ctx, &ssm.GetParametersByPathInput{Path: &path})
	if err != nil {
		err = fmt.Errorf("failed to load SSM parameters by path [%s]: %w", path, err)
		return []types.Parameter{}, err
	}

	return ssmOutput.Parameters, nil
}

func extractParameter(params []types.Parameter, key string) (string, error) {
	paramSuffix := fmt.Sprintf("/%s", key)

	for _, param := range params {
		if strings.HasSuffix(*param.Name, paramSuffix) {
			return *param.Value, nil
		}
	}

	return "", fmt.Errorf("parameter [%s] not found", key)
}
