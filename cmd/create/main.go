package main

import (
	"os"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws/external"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/projects/secure-notes/internal/creating"
	"github.com/projects/secure-notes/internal/http/rest"
	"github.com/projects/secure-notes/internal/platform/security"
	"github.com/projects/secure-notes/internal/platform/web"
	db "github.com/projects/secure-notes/internal/storage/dynamodb"
	"go.uber.org/zap"
)

var createNoteHandler web.Handler

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
	now := func() time.Time { return time.Now().UTC() }
	hashGen := security.GenerateHashWithSalt
	creator := creating.NewService(storage, now, hashGen)

	// setup API Gateway adapter
	handler := rest.CreateNote(creator)

	// setup web middleware
	logger, err := zap.NewProductionConfig().Build()
	if err != nil {
		panic("cannot initialize logger")
	}
	middleware := &web.Middleware{
		Logger: logger.Sugar(),
	}

	// setup handler wrapped with middleware
	createNoteHandler = middleware.WrapWithCorsAndLogging(handler)
}

func main() {
	lambda.Start(createNoteHandler)
}
