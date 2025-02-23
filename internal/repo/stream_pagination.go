package repo

import (
	"fmt"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"strings"
)

type StreamNextPageKey struct {
	Pk         string `dynamodbav:"PK"`
	Sk         int    `dynamodbav:"SK"`
	StreamType string
	UpdatedAt  string
}

func ParseNextPageKey(nextPageKey string) (map[string]types.AttributeValue, error) {
	parts := strings.Split(nextPageKey, "|")
	streamNextPageKey := StreamNextPageKey{
		Pk:         parts[0],
		Sk:         0,
		StreamType: parts[1],
		UpdatedAt:  parts[2],
	}

	key, err := attributevalue.MarshalMap(streamNextPageKey)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal next page key: %w", err)
	}

	return key, nil
}

func FormatNextPageKey(lastEvaluatedKey map[string]types.AttributeValue) (string, error) {
	var streamNextPageKey StreamNextPageKey
	err := attributevalue.UnmarshalMap(lastEvaluatedKey, &streamNextPageKey)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal next page key: %w", err)
	}

	parts := []string{
		streamNextPageKey.Pk,
		streamNextPageKey.StreamType,
		streamNextPageKey.UpdatedAt,
	}

	return strings.Join(parts, "|"), nil
}
