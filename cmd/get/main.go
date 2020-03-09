package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/external"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/dynamodbattribute"
	"github.com/repos/secure-notes/internal/web"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

var (
	dbCli      *dynamodb.Client
	middleware *web.Middleware
)

type note struct {
	ID   string `json:"id"`
	Text string `json:"text"`
	TTL  int64  `json:"ttl"`
}

type secureNote struct {
	ID          string `dynamodbav:"pk"`
	Text        string `dynamodbav:"text"`
	Hash        string `dynamodbav:"hash"`
	TTL         int64  `dynamodbav:"ttl"`
	OneTimeRead bool   `dynamodbav:"oneTimeRead"`
}

var (
	errNoteNotFound = errors.New("note not found")
)

func init() {
	cfg, err := external.LoadDefaultAWSConfig()
	if err != nil {
		panic("cannot read AWS configuration")
	}

	dbCli = dynamodb.New(cfg)

	logger, err := zap.NewProductionConfig().Build()
	if err != nil {
		panic("cannot initialize logger")
	}
	middleware = &web.Middleware{
		Logger: logger.Sugar(),
	}
}

func Handler(ctx context.Context, req web.Request) (web.Response, error) {
	noteID := req.PathParameters["id"]

	secNote, err := get(ctx, dbCli, noteID)
	if err != nil {
		switch err {
		case errNoteNotFound:
			return web.Response{
				StatusCode: http.StatusNotFound,
			}, fmt.Errorf("get note from db: %w", err)
		default:
			return web.InternalServerError(), fmt.Errorf("get note from db: %w", err)
		}
	}

	plainPwd := req.Headers["password"]
	ok := comparePasswords(secNote.Hash, []byte(plainPwd))
	if !ok {
		return web.Response{
			StatusCode: http.StatusUnauthorized,
		}, fmt.Errorf("wrong password")
	}

	resp, err := createResponse(secNote, err)
	if err != nil {
		return web.InternalServerError(), fmt.Errorf("create response: %w", err)
	}

	if secNote.OneTimeRead {
		if err := delete(noteID, ctx); err != nil {
			return web.InternalServerError(), fmt.Errorf("delete note: %w", err)
		}
	}

	return resp, nil
}

func get(ctx context.Context, dbCli *dynamodb.Client, noteID string) (secureNote, error) {
	input := dynamodb.GetItemInput{
		Key: map[string]dynamodb.AttributeValue{
			"pk": {
				S: aws.String(noteID),
			},
		},
		TableName: aws.String("notes"),
	}

	item, err := dbCli.GetItemRequest(&input).Send(ctx)
	if err != nil {
		return secureNote{}, fmt.Errorf("get item from db: %w", err)
	}

	if notFound := len(item.Item) == 0; notFound {
		return secureNote{}, errNoteNotFound
	}

	var secNote secureNote
	if err := dynamodbattribute.UnmarshalMap(item.Item, &secNote); err != nil {
		return secureNote{}, fmt.Errorf("unmarshal note from db map: %w", err)
	}

	return secNote, nil
}

func createResponse(secNote secureNote, err error) (web.Response, error) {
	n := note{
		ID:   secNote.ID,
		Text: secNote.Text,
		TTL:  secNote.TTL,
	}

	noteBytes, err := json.Marshal(n)
	if err != nil {
		return web.Response{}, fmt.Errorf("json marshal response: %w", err)
	}

	resp := web.Response{
		StatusCode: http.StatusOK,
		Body:       string(noteBytes),
	}
	return resp, nil
}

func delete(noteID string, ctx context.Context) error {
	input := dynamodb.DeleteItemInput{
		Key: map[string]dynamodb.AttributeValue{
			"pk": {
				S: aws.String(noteID),
			},
		},
		TableName: aws.String("notes"),
	}
	if _, err := dbCli.DeleteItemRequest(&input).Send(ctx); err != nil {
		return fmt.Errorf("delete note from db: %w", err)
	}
	return nil
}

func comparePasswords(hashedPwd string, plainPwd []byte) bool {
	byteHash := []byte(hashedPwd)
	err := bcrypt.CompareHashAndPassword(byteHash, plainPwd)
	if err != nil {
		return false
	}

	return true
}

func main() {
	lambda.Start(middleware.WrapWithCorsAndLogging(Handler))
}
