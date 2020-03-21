package dynamodb

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/dynamodbattribute"
	"github.com/projects/secure-notes/internal/creating"
)

type Storage struct {
	DbCli     *dynamodb.Client
	TableName string
}

func NewStorage(dbCli *dynamodb.Client, tableName string) *Storage {
	return &Storage{
		DbCli:     dbCli,
		TableName: tableName,
	}
}

func (s *Storage) CreateNote(ctx context.Context, sn creating.SecureNote) error {
	newNote := Note{
		ID:          sn.ID,
		Text:        sn.Text,
		Hash:        sn.Hash,
		TTL:         sn.TTL,
		OneTimeRead: sn.OneTimeRead,
	}

	item, err := dynamodbattribute.MarshalMap(newNote)
	if err != nil {
		return fmt.Errorf("marshal note to db map: %w", err)
	}

	input := dynamodb.PutItemInput{
		Item:      item,
		TableName: aws.String(s.TableName),
	}
	if _, err := s.DbCli.PutItemRequest(&input).Send(ctx); err != nil {
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
		TableName:        aws.String(s.TableName),
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
