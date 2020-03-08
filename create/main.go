package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/dynamodbattribute"
	"golang.org/x/crypto/bcrypt"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws/external"
	"github.com/google/uuid"
)

var dbCli *dynamodb.Client

type Request events.APIGatewayProxyRequest
type Response events.APIGatewayProxyResponse

//TODO validate body
//TODO create human friendly urls
//TODO think about encoding the note to base64 because of chinese characters
//TODO button for nice url copy
//TODO create page if note expired
//TODO destroy note with button after unlock

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

	securedNote := newSecureNote(n)
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

func newSecureNote(n note) secureNote {
	now := time.Now().UTC()
	ttl := now.Add(time.Duration(n.LifeTimeSeconds) * time.Second).Unix()

	return secureNote{
		ID:   uuid.New().String(),
		Text: n.Text,
		Hash: hashAndSalt([]byte(n.Password)),
		TTL:  ttl,
	}
}

func hashAndSalt(pwd []byte) string {
	hash, err := bcrypt.GenerateFromPassword(pwd, bcrypt.MinCost)
	if err != nil {
		log.Println(err)
	}

	return string(hash)
}

func main() {
	lambda.Start(Handler)
}
