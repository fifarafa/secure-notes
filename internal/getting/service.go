package getting

import (
	"context"
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

var (
	// ErrNotFound is used when a beer could not be found.
	ErrNotFound = errors.New("note not found")

	// ErrNotAuthorized
	ErrNotAuthorized = errors.New("wrong password")
)

type Service struct {
	repository
}

type repository interface {
	GetNote(ctx context.Context, noteID string) (SecureNote, error)
	DeleteNote(ctx context.Context, noteID string) error
}

func NewService(repository repository) *Service {
	return &Service{repository: repository}
}

func (s *Service) Get(ctx context.Context, noteID, password string) (Note, error) {
	secureNote, err := s.repository.GetNote(ctx, noteID)
	if err != nil {
		return Note{}, fmt.Errorf("repository get note: %w", err)
	}

	ok := verifyPassword(secureNote.Hash, password)
	if !ok {
		return Note{}, ErrNotAuthorized
	}

	if secureNote.OneTimeRead {
		if err := s.repository.DeleteNote(ctx, secureNote.ID); err != nil {
			return Note{}, fmt.Errorf("delete note: %w", err)
		}
	}

	return Note{
		ID:   secureNote.ID,
		Text: secureNote.Text,
		TTL:  secureNote.TTL,
	}, nil
}

func verifyPassword(hashedPwd, plainPwd string) bool {
	byteHash := []byte(hashedPwd)
	bytePwd := []byte(plainPwd)
	err := bcrypt.CompareHashAndPassword(byteHash, bytePwd)
	if err != nil {
		return false
	}

	return true
}
