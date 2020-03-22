package main

import (
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/projects/secure-notes/internal/getting"
	"github.com/projects/secure-notes/internal/http/rest"
	"github.com/projects/secure-notes/internal/platform/provider"
	"github.com/projects/secure-notes/internal/platform/web"
)

var getNoteHandler web.Handler

func init() {
	cfg := provider.AWSConfig()
	storage := provider.DynamoStorage(cfg, os.Getenv("NOTES_TABLE"))
	getter := getting.NewService(storage)
	handler := rest.GetNote(getter)
	middleware := provider.Middleware()
	getNoteHandler = middleware.WrapWithCorsAndLogging(handler)
}

func main() {
	lambda.Start(getNoteHandler)
}
