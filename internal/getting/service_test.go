package getting_test

import (
	"context"
	"testing"
	"time"

	"github.com/projects/secure-notes/internal/getting"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestService_GetNoteOK(t *testing.T) {
	// given
	repository := mockRepository{}
	repository.On("GetNote", "qx2rx").Return(getting.SecureNote{
		ID:          "qx2rx",
		Text:        "Hello World",
		Hash:        "$2a$04$tD4EmWTb6FficqPruQNzL.t4X79mud7a3ybAp6JYgf7fItsw3pRoC",
		TTL:         time.Date(2020, 3, 22, 16, 0, 0, 0, time.UTC).Unix(),
		OneTimeRead: true,
	}, nil)
	repository.On("DeleteNote", "qx2rx").Return(nil)

	s := getting.NewService(&repository)

	// when
	gotNote, gotErr := s.GetNote(context.TODO(), "qx2rx", "abc")

	// then
	assert.NoError(t, gotErr)
	assert.Equal(t, getting.Note{
		ID:   "qx2rx",
		Text: "Hello World",
		TTL:  time.Date(2020, 3, 22, 16, 0, 0, 0, time.UTC).Unix(),
	}, gotNote)
}

func TestService_GetNoteWrongPassword(t *testing.T) {
	// given
	repository := mockRepository{}
	repository.On("GetNote", "qx2rx").Return(getting.SecureNote{
		ID:          "qx2rx",
		Text:        "Hello World",
		Hash:        "$2a$04$tD4EmWTb6FficqPruQNzL.t4X79mud7a3ybAp6JYgf7fItsw3pRoC",
		TTL:         time.Date(2020, 3, 22, 16, 0, 0, 0, time.UTC).Unix(),
		OneTimeRead: true,
	}, nil)

	s := getting.NewService(&repository)

	// when
	gotNote, gotErr := s.GetNote(context.TODO(), "qx2rx", "wrongpassword")

	// then
	assert.EqualError(t, gotErr, "wrong password")
	assert.Equal(t, getting.Note{}, gotNote)
}

func TestService_GetNoteNotExists(t *testing.T) {
	// given
	repository := mockRepository{}
	repository.On("GetNote", "qx2rx").Return(getting.SecureNote{}, getting.ErrNotFound)

	s := getting.NewService(&repository)

	// when
	gotNote, gotErr := s.GetNote(context.TODO(), "qx2rx", "abc")

	// then
	assert.EqualError(t, gotErr, "repository get note: note not found")
	assert.Equal(t, getting.Note{}, gotNote)
}

type mockRepository struct {
	mock.Mock
}

func (m *mockRepository) GetNote(ctx context.Context, noteID string) (getting.SecureNote, error) {
	args := m.Called(noteID)
	return args.Get(0).(getting.SecureNote), args.Error(1)
}

func (m *mockRepository) DeleteNote(ctx context.Context, noteID string) error {
	args := m.Called(noteID)
	return args.Error(0)
}
