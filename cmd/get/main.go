package main

import (
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws/external"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/projects/secure-notes/internal/getting"
	"github.com/projects/secure-notes/internal/http/rest"
	"github.com/projects/secure-notes/internal/platform/web"
	db "github.com/projects/secure-notes/internal/storage/dynamodb"
	"go.uber.org/zap"
)

var getNoteHandler web.Handler

func init() {
	// load AWS config
	cfg, err := external.LoadDefaultAWSConfig()
	if err != nil {
		panic("cannot read AWS configuration")
	}

	// setup DynamoDB adapter
	dbCli := dynamodb.New(cfg)
	notesTableName := os.Getenv("NOTES_TABLE")
	storage := db.NewStorage(dbCli, notesTableName)

	// setup domain service
	getter := getting.NewService(storage)

	// setup API Gateway adapter
	handler := rest.GetNote(getter)

	// setup web middleware
	logger, err := zap.NewProductionConfig().Build()
	if err != nil {
		panic("cannot initialize logger")
	}
	middleware := &web.Middleware{
		Logger: logger.Sugar(),
	}

	// setup handler wrapped with middleware
	getNoteHandler = middleware.WrapWithCorsAndLogging(handler)
}

func main() {
	lambda.Start(getNoteHandler)
}
