package creating_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/projects/secure-notes/internal/creating"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestService_CreateNoteOK(t *testing.T) {
	// given
	createNote := creating.Note{
		Text:            "Hello World",
		Password:        "abc",
		LifeTimeSeconds: 3600,
		OneTimeRead:     true,
	}

	repository := mockRepository{}
	repository.On("IncrementNoteCounter").Return(1, nil)
	repository.On("CreateNote", creating.SecureNote{
		ID:          "qx2rx",
		Text:        "Hello World",
		Hash:        "$2a$04$tD4EmWTb6FficqPruQNzL.t4X79mud7a3ybAp6JYgf7fItsw3pRoC",
		TTL:         time.Date(2020, 3, 22, 16, 0, 0, 0, time.UTC).Unix(),
		OneTimeRead: true,
	}).Return(nil)

	timer := func() time.Time {
		return time.Date(2020, 3, 22, 15, 0, 0, 0, time.UTC)
	}

	hashGen := func(pwd string) (string, error) {
		return "$2a$04$tD4EmWTb6FficqPruQNzL.t4X79mud7a3ybAp6JYgf7fItsw3pRoC", nil
	}

	// when
	s := creating.NewService(&repository, timer, hashGen)

	// then
	gotNoteID, gotErr := s.CreateNote(context.TODO(), createNote)

	assert.NoError(t, gotErr)
	assert.Equal(t, "qx2rx", gotNoteID)
}

func TestService_CreateNote(t *testing.T) {
	// given
	createNote := creating.Note{
		Text:            "Hello World",
		Password:        "abc",
		LifeTimeSeconds: 3600,
		OneTimeRead:     true,
	}

	repository := mockRepository{}
	repository.On("IncrementNoteCounter").Return(1, nil)
	repository.On("CreateNote", creating.SecureNote{
		ID:          "qx2rx",
		Text:        "Hello World",
		Hash:        "$2a$04$tD4EmWTb6FficqPruQNzL.t4X79mud7a3ybAp6JYgf7fItsw3pRoC",
		TTL:         time.Date(2020, 3, 22, 16, 0, 0, 0, time.UTC).Unix(),
		OneTimeRead: true,
	}).Return(errors.New("some error from database"))

	timer := func() time.Time {
		return time.Date(2020, 3, 22, 15, 0, 0, 0, time.UTC)
	}

	hashGen := func(pwd string) (string, error) {
		return "$2a$04$tD4EmWTb6FficqPruQNzL.t4X79mud7a3ybAp6JYgf7fItsw3pRoC", nil
	}

	// when
	s := creating.NewService(&repository, timer, hashGen)

	// then
	gotNoteID, gotErr := s.CreateNote(context.TODO(), createNote)

	assert.EqualError(t, gotErr, "repository create secured note: some error from database")
	assert.Equal(t, "", gotNoteID)
}

type mockRepository struct {
	mock.Mock
}

func (m *mockRepository) CreateNote(ctx context.Context, sn creating.SecureNote) error {
	args := m.Called(sn)
	return args.Error(0)
}

func (m *mockRepository) IncrementNoteCounter(ctx context.Context) (int, error) {
	args := m.Called()
	return args.Int(0), args.Error(1)
}
