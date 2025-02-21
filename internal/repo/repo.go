package repo

import (
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

type EsRepo struct {
	dynamoDb  *dynamodb.Client
	tableName string
}

func NewEsRepo(dynamoDb *dynamodb.Client, tableName string) *EsRepo {
	return &EsRepo{
		dynamoDb:  dynamoDb,
		tableName: tableName,
	}
}
