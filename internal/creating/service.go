package creating

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/speps/go-hashids"
	"golang.org/x/crypto/bcrypt"
)

// Service provides note creating operation
type Service struct {
	repository
}

func NewService(repository repository) *Service {
	return &Service{repository: repository}
}

type repository interface {
	CreateNote(context.Context, SecureNote) error
	IncrementNoteCounter(context.Context) (int, error)
}

// CreateNote creates secure note in storage
func (s *Service) CreateNote(ctx context.Context, plain Note) (noteID string, err error) {
	now := time.Now().UTC()
	noteTTL := now.Add(time.Duration(plain.LifeTimeSeconds) * time.Second).Unix()

	saltedHash, err := generateHashWithSalt([]byte(plain.Password))
	if err != nil {
		return "", fmt.Errorf("generate hash with salt: %w", err)
	}

	counter, err := s.IncrementNoteCounter(ctx)
	if err != nil {
		return "", fmt.Errorf("increment note counter: %w", err)
	}

	id := generateHumanFriendlyID(counter)

	securedNote := SecureNote{
		ID:          id,
		Text:        plain.Text,
		Hash:        saltedHash,
		TTL:         noteTTL,
		OneTimeRead: plain.OneTimeRead,
	}

	if err := s.repository.CreateNote(ctx, securedNote); err != nil {
		return "", fmt.Errorf("storage create secured note: %w", err)
	}

	return "", nil
}

func generateHashWithSalt(pwd []byte) (string, error) {
	hash, err := bcrypt.GenerateFromPassword(pwd, bcrypt.MinCost)
	if err != nil {
		return "", errors.New("bcrypt generate from password")
	}

	return string(hash), nil
}

func generateHumanFriendlyID(noteCounter int) string {
	hd := hashids.NewData()
	hd.Salt = "salt for secure notes app"
	hd.MinLength = 5
	h, _ := hashids.NewWithData(hd)
	e, _ := h.Encode([]int{noteCounter})
	return e
}
