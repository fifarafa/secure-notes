package creating

import (
	"context"
	"fmt"
	"time"

	"github.com/speps/go-hashids"
)

// Service provides note creating operation
type Service struct {
	repo            repository
	now             func() time.Time
	genHashWithSalt func(password string) (string, error)
}

type repository interface {
	CreateNote(context.Context, SecureNote) error
	IncrementNoteCounter(context.Context) (int, error)
}

// NewService provides creating note service
func NewService(r repository, now func() time.Time, genHashWithSalt func(password string) (string, error)) *Service {
	return &Service{repo: r, now: now, genHashWithSalt: genHashWithSalt}
}

// CreateNote creates secure note in storage
func (s *Service) CreateNote(ctx context.Context, plain Note) (noteID string, err error) {
	noteTTL := s.now().Add(time.Duration(plain.LifeTimeSeconds) * time.Second).Unix()

	saltedHash, err := s.genHashWithSalt(plain.Password)
	if err != nil {
		return "", fmt.Errorf("generate hash with salt: %w", err)
	}

	counter, err := s.repo.IncrementNoteCounter(ctx)
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

	if err := s.repo.CreateNote(ctx, securedNote); err != nil {
		return "", fmt.Errorf("repository create secured note: %w", err)
	}

	return securedNote.ID, nil
}

func generateHumanFriendlyID(noteCounter int) string {
	hd := hashids.NewData()
	hd.Salt = "salt for secure notes app"
	hd.MinLength = 5
	h, _ := hashids.NewWithData(hd)
	e, _ := h.Encode([]int{noteCounter})
	return e
}
