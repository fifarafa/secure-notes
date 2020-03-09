package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/external"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/dynamodbattribute"
	"github.com/google/uuid"
	"github.com/repos/secure-notes/internal/web"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

var (
	dbCli      *dynamodb.Client
	middleware *web.Middleware
)

//TODO create human friendly urls

type note struct {
	Text            string `json:"text"`
	Password        string `json:"password"`
	LifeTimeSeconds int64  `json:"lifeTimeSeconds"`
	OneTimeRead     bool   `json:"oneTimeRead"`
}

type secureNote struct {
	ID          string `dynamodbav:"pk"`
	Text        string `dynamodbav:"text"`
	Hash        string `dynamodbav:"hash"`
	TTL         int64  `dynamodbav:"ttl"`
	OneTimeRead bool   `dynamodbav:"oneTimeRead"`
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
	var n note
	if err := json.Unmarshal([]byte(req.Body), &n); err != nil {
		return web.Response{
			StatusCode: http.StatusBadRequest,
		}, err
	}

	secNote, err := newSecureNote(n)
	if err != nil {
		return web.InternalServerError(), fmt.Errorf("new secure note: %w", err)
	}

	if err := save(ctx, secNote); err != nil {
		return web.InternalServerError(), fmt.Errorf("save secured note: %w", err)
	}

	resp, err := createResponse(secNote.ID)
	if err != nil {
		return web.InternalServerError(), fmt.Errorf("create response: %w", err)
	}

	return resp, nil
}

func save(ctx context.Context, sn secureNote) error {
	item, err := dynamodbattribute.MarshalMap(sn)
	if err != nil {
		return fmt.Errorf("marshal note to db map: %w", err)
	}

	input := dynamodb.PutItemInput{
		Item:      item,
		TableName: aws.String("notes"),
	}
	if _, err := dbCli.PutItemRequest(&input).Send(ctx); err != nil {
		log.Print(err)
		return fmt.Errorf("put item in db: %w", err)
	}

	return nil
}

func newSecureNote(n note) (secureNote, error) {
	now := time.Now().UTC()
	ttl := now.Add(time.Duration(n.LifeTimeSeconds) * time.Second).Unix()
	saltedHash, err := generateHashWithSalt([]byte(n.Password))
	if err != nil {
		return secureNote{}, fmt.Errorf("generate hash with salt: %w", err)
	}

	return secureNote{
		ID:          uuid.New().String(),
		Text:        n.Text,
		Hash:        saltedHash,
		TTL:         ttl,
		OneTimeRead: n.OneTimeRead,
	}, nil
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
