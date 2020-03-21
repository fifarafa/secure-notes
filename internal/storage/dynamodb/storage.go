package dynamodb

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/dynamodbattribute"
	"github.com/projects/secure-notes/internal/creating"
	"github.com/projects/secure-notes/internal/getting"
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

func (s *Storage) GetNote(ctx context.Context, noteID string) (getting.SecureNote, error) {
	input := dynamodb.GetItemInput{
		Key: map[string]dynamodb.AttributeValue{
			"pk": {
				S: aws.String(noteID),
			},
		},
		TableName: aws.String(s.TableName),
	}

	item, err := s.DbCli.GetItemRequest(&input).Send(ctx)
	if err != nil {
		return getting.SecureNote{}, fmt.Errorf("get item from db: %w", err)
	}

	if notFound := len(item.Item) == 0; notFound {
		return getting.SecureNote{}, getting.ErrNotFound
	}

	var n Note
	if err := dynamodbattribute.UnmarshalMap(item.Item, &n); err != nil {
		return getting.SecureNote{}, fmt.Errorf("unmarshal note from db map: %w", err)
	}

	note := getting.SecureNote{
		ID:          n.ID,
		Text:        n.Text,
		Hash:        n.Hash,
		TTL:         n.TTL,
		OneTimeRead: n.OneTimeRead,
	}

	return note, nil
}

func (s *Storage) DeleteNote(ctx context.Context, noteID string) error {
	input := dynamodb.DeleteItemInput{
		Key: map[string]dynamodb.AttributeValue{
			"pk": {
				S: aws.String(noteID),
			},
		},
		TableName: aws.String(s.TableName),
	}
	if _, err := s.DbCli.DeleteItemRequest(&input).Send(ctx); err != nil {
		return fmt.Errorf("delete note from db: %w", err)
	}

	return nil
}
