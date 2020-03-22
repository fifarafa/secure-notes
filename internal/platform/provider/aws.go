package provider

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/external"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/projects/secure-notes/internal/platform/web"
	db "github.com/projects/secure-notes/internal/storage/dynamodb"
	"go.uber.org/zap"
)

func AWSConfig() aws.Config {
	cfg, err := external.LoadDefaultAWSConfig()
	if err != nil {
		panic("cannot read AWS configuration")
	}

	return cfg
}

func DynamoStorage(cfg aws.Config, tableName string) *db.Storage {
	dbCli := dynamodb.New(cfg)
	storage := db.NewStorage(dbCli, tableName)
	return storage
}

func Middleware() *web.Middleware {
	logger, err := zap.NewProductionConfig().Build()
	if err != nil {
		panic("cannot initialize logger")
	}
	middleware := web.Middleware{
		Logger: logger.Sugar(),
	}
	return &middleware
}
