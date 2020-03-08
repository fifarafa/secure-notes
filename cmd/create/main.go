package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/external"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/dynamodbattribute"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var dbCli *dynamodb.Client

type Request events.APIGatewayProxyRequest
type Response events.APIGatewayProxyResponse

//TODO validate body
//TODO create human friendly urls
//TODO create page if note expired???
//TODO destroy note tick after first read IDEA

//TODO divide into cmd
//TODO create internal

type note struct {
	Text            string `json:"text"`
	Password        string `json:"password"`
	LifeTimeSeconds int64  `json:"lifeTimeSeconds"`
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
}

// Handler is our lambda handler invoked by the `lambda.Start` function call
func Handler(ctx context.Context, req Request) (Response, error) {
	var n note
	if err := json.Unmarshal([]byte(req.Body), &n); err != nil {
		log.Print(err)
		return Response{
			Headers: map[string]string{
				"Access-Control-Allow-Origin":      "*",
				"Access-Control-Allow-Credentials": "true",
			},
			StatusCode: http.StatusBadRequest,
		}, nil
	}

	securedNote, err := newSecureNote(n)
	if err != nil {
		log.Print(err)
		return Response{
			Headers: map[string]string{
				"Access-Control-Allow-Origin":      "*",
				"Access-Control-Allow-Credentials": "true",
			},
			StatusCode: http.StatusInternalServerError,
		}, nil
	}

	item, err := dynamodbattribute.MarshalMap(securedNote)
	if err != nil {
		log.Print(err)
		return Response{
			Headers: map[string]string{
				"Access-Control-Allow-Origin":      "*",
				"Access-Control-Allow-Credentials": "true",
			},
			StatusCode: http.StatusInternalServerError,
		}, nil
	}

	input := dynamodb.PutItemInput{
		Item:      item,
		TableName: aws.String("notes"),
	}
	if _, err := dbCli.PutItemRequest(&input).Send(ctx); err != nil {
		log.Print(err)
		return Response{
			Headers: map[string]string{
				"Access-Control-Allow-Origin":      "*",
				"Access-Control-Allow-Credentials": "true",
			},
			StatusCode: http.StatusInternalServerError,
		}, nil
	}

	type ResponseId struct {
		ID string `json:"id"`
	}

	data, err := json.Marshal(&ResponseId{ID: securedNote.ID})
	if err != nil {
		return Response{
			Headers: map[string]string{
				"Access-Control-Allow-Origin":      "*",
				"Access-Control-Allow-Credentials": "true",
			},
			StatusCode: http.StatusInternalServerError,
		}, nil
	}

	resp := Response{
		StatusCode: http.StatusCreated,
		Body:       string(data),
		Headers: map[string]string{
			"Access-Control-Allow-Origin":      "*",
			"Access-Control-Allow-Credentials": "true",
		},
	}

	return resp, nil
}

func newSecureNote(n note) (*secureNote, error) {
	now := time.Now().UTC()
	ttl := now.Add(time.Duration(n.LifeTimeSeconds) * time.Second).Unix()
	saltedHash, err := generateHashWithSalt([]byte(n.Password))
	if err != nil {
		return nil, fmt.Errorf("generate hash with salt: %w", err)
	}

	return &secureNote{
		ID:   uuid.New().String(),
		Text: n.Text,
		Hash: saltedHash,
		TTL:  ttl,
	}, nil
}

func generateHashWithSalt(pwd []byte) (string, error) {
	hash, err := bcrypt.GenerateFromPassword(pwd, bcrypt.MinCost)
	if err != nil {
		return "", errors.New("bcrypt generate from password")
	}

	return string(hash), nil
}

func main() {
	lambda.Start(Handler)
}
