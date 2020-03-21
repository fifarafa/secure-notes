package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws/external"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/projects/secure-notes/internal/creating"
	"github.com/projects/secure-notes/internal/web"
	"github.com/speps/go-hashids"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

var (
	dbCli          *dynamodb.Client
	middleware     *web.Middleware
	notesTableName string
)

func init() {
	cfg, err := external.LoadDefaultAWSConfig()
	if err != nil {
		panic("cannot read AWS configuration")
	}

	dbCli = dynamodb.New(cfg)
	notesTableName = os.Getenv("NOTES_TABLE")

	logger, err := zap.NewProductionConfig().Build()
	if err != nil {
		panic("cannot initialize logger")
	}
	middleware = &web.Middleware{
		Logger: logger.Sugar(),
	}
}

func Handler(s creating.Service) func(ctx context.Context, req web.Request) (web.Response, error) {
	return func(ctx context.Context, req web.Request) (web.Response, error) {
		var newNote creating.Note
		if err := json.Unmarshal([]byte(req.Body), &n); err != nil {
			return web.Response{
				StatusCode: http.StatusBadRequest,
			}, err
		}

		noteID := s.CreateNote(newNote)

		resp, err := createResponse(noteID)
		if err != nil {
			return web.InternalServerError(), fmt.Errorf("create response: %w", err)
		}

		return resp, nil
	}
}

func generateHumanFriendlyID(noteCounter int) string {
	hd := hashids.NewData()
	hd.Salt = "salt for secure notes app"
	hd.MinLength = 5
	h, _ := hashids.NewWithData(hd)
	e, _ := h.Encode([]int{noteCounter})
	return e
}

func createResponse(noteID string) (web.Response, error) {
	type ResponseId struct {
		ID string `json:"id"`
	}

	responseBytes, err := json.Marshal(&ResponseId{ID: noteID})
	if err != nil {
		return web.Response{}, fmt.Errorf("json marshal response: %w", err)
	}

	resp := web.Response{
		StatusCode: http.StatusCreated,
		Body:       string(responseBytes),
	}
	return resp, nil
}

func generateHashWithSalt(pwd []byte) (string, error) {
	hash, err := bcrypt.GenerateFromPassword(pwd, bcrypt.MinCost)
	if err != nil {
		return "", errors.New("bcrypt generate from password")
	}

	return string(hash), nil
}

func main() {
	lambda.Start(middleware.WrapWithCorsAndLogging(Handler))
}
