package dynamodb

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/dynamodbattribute"
)

type Storage struct {
	DbCli     *dynamodb.Client
	TableName string
}

func (s *Storage) Save(ctx context.Context, sn secureNote) error {
	item, err := dynamodbattribute.MarshalMap(sn)
	if err != nil {
		return fmt.Errorf("marshal note to db map: %w", err)
	}

	input := dynamodb.PutItemInput{
		Item:      item,
		TableName: aws.String(notesTableName),
	}
	if _, err := dbCli.PutItemRequest(&input).Send(ctx); err != nil {
		log.Print(err)
		return fmt.Errorf("put item in db: %w", err)
	}

	return nil
}

func (s *Storage) IncrementNoteCounter(ctx context.Context) (int, error) {
	input := dynamodb.UpdateItemInput{
		ExpressionAttributeNames: map[string]string{
			"#counter": "counter",
		},
		ExpressionAttributeValues: map[string]dynamodb.AttributeValue{
			":n": {
				N: aws.String("1"),
			},
		},
		Key: map[string]dynamodb.AttributeValue{
			"pk": {
				S: aws.String("__id"),
			},
		},
		ReturnValues:     "UPDATED_NEW",
		TableName:        aws.String(notesTableName),
		UpdateExpression: aws.String("add #counter :n"),
	}

	resp, err := s.DbCli.UpdateItemRequest(&input).Send(ctx)
	if err != nil {
		return 0, fmt.Errorf("update note counter: %w", err)
	}

	type counter struct {
		Counter int `dynamodbav:"counter"`
	}
	var c counter
	if err := dynamodbattribute.UnmarshalMap(resp.UpdateItemOutput.Attributes, &c); err != nil {
		return 0, fmt.Errorf("unmarshal counter from db map: %w", err)
	}

	return c.Counter, nil
}