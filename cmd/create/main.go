package main

import (
	"os"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/projects/secure-notes/internal/creating"
	"github.com/projects/secure-notes/internal/http/rest"
	"github.com/projects/secure-notes/internal/platform/provider"
	"github.com/projects/secure-notes/internal/platform/security"
	"github.com/projects/secure-notes/internal/platform/web"
)

var createNoteHandler web.Handler

func init() {
	cfg := provider.AWSConfig()
	storage := provider.DynamoStorage(cfg, os.Getenv("NOTES_TABLE"))

	now := func() time.Time { return time.Now().UTC() }
	hashGen := security.GenerateHashWithSalt
	creator := creating.NewService(storage, now, hashGen)

	handler := rest.CreateNote(creator)
	middleware := provider.Middleware()
	createNoteHandler = middleware.WrapWithCorsAndLogging(handler)
}

func main() {
	lambda.Start(createNoteHandler)
}
