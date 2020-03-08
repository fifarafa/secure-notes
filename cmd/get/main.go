package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
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
	ID   string `dynamodbav:"pk"`
	Text string `dynamodbav:"text"`
	Hash string `dynamodbav:"hash"`
	TTL  int64  `dynamodbav:"ttl"`
}

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
		log.Print(err)
		return web.Response{
			StatusCode: http.StatusInternalServerError,
			Headers: map[string]string{
				"Access-Control-Allow-Origin":      "*",
				"Access-Control-Allow-Credentials": "true",
			},
		}, nil
	}

	if notFound := len(item.Item) == 0; notFound {
		log.Print("empty item")
		return web.Response{
			StatusCode: http.StatusNotFound,
			Headers: map[string]string{
				"Access-Control-Allow-Origin":      "*",
				"Access-Control-Allow-Credentials": "true",
			},
		}, nil
	}

	var secureNote secureNote
	if err := dynamodbattribute.UnmarshalMap(item.Item, &secureNote); err != nil {
		log.Print(err)
		return web.Response{
			StatusCode: http.StatusInternalServerError,
			Headers: map[string]string{
				"Access-Control-Allow-Origin":      "*",
				"Access-Control-Allow-Credentials": "true",
			},
		}, err
	}

	//TODO validate
	plainPwd := req.Headers["note-secret"]

	ok, err := comparePasswords(secureNote.Hash, []byte(plainPwd))
	if err != nil {
		log.Print(err)
		return web.Response{
			StatusCode: http.StatusInternalServerError,
			Headers: map[string]string{
				"Access-Control-Allow-Origin":      "*",
				"Access-Control-Allow-Credentials": "true",
			},
		}, err
	}

	if !ok {
		return web.Response{
			StatusCode: http.StatusUnauthorized,
			Headers: map[string]string{
				"Access-Control-Allow-Origin":      "*",
				"Access-Control-Allow-Credentials": "true",
			},
		}, err
	}

	n := note{
		ID:   secureNote.ID,
		Text: secureNote.Text,
		TTL:  secureNote.TTL,
	}

	data, err := json.Marshal(n)
	if err != nil {
		log.Print(err)
		return web.Response{
			StatusCode: http.StatusInternalServerError,
			Headers: map[string]string{
				"Access-Control-Allow-Origin":      "*",
				"Access-Control-Allow-Credentials": "true",
			},
		}, nil
	}

	resp := web.Response{
		StatusCode: http.StatusOK,
		Body:       string(data),
		Headers: map[string]string{
			"Access-Control-Allow-Origin":      "*",
			"Access-Control-Allow-Credentials": "true",
		},
	}

	return resp, nil
}

func comparePasswords(hashedPwd string, plainPwd []byte) (bool, error) {
	byteHash := []byte(hashedPwd)
	err := bcrypt.CompareHashAndPassword(byteHash, plainPwd)
	if err != nil {
		return false, errors.New("bcrypt compare hash with password")
	}

	return true, nil
}

func main() {
	lambda.Start(middleware.WrapWithCorsAndLogging(Handler))
}
