package rest_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/projects/secure-notes/internal/creating"
	"github.com/projects/secure-notes/internal/http/rest"
	"github.com/projects/secure-notes/internal/platform/web"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_CreateNoteOK(t *testing.T) {
	// given
	service := mockCreateService{}
	service.On("CreateNote", creating.Note{
		Text:            "Hello World",
		Password:        "mySecretPassword",
		LifeTimeSeconds: 360000,
		OneTimeRead:     true,
	}).Return("qx2rx", nil)

	handler := rest.CreateNote(&service)

	request := web.Request{
		Body: `{
				"text": "Hello World",
				"lifeTimeSeconds": 360000,
				"password": "mySecretPassword",
				"oneTimeRead": true
			}`,
	}

	// when
	gotResp, gotErr := handler(context.TODO(), request)

	// then
	assert.Equal(t, web.Response{
		StatusCode: http.StatusCreated,
		Body:       `{"id":"qx2rx"}`,
	}, gotResp)
	assert.NoError(t, gotErr)
}

func Test_CreateNoteInternalError(t *testing.T) {
	// given
	service := mockCreateService{}
	service.On("CreateNote", creating.Note{
		Text:            "Hello World",
		Password:        "mySecretPassword",
		LifeTimeSeconds: 360000,
		OneTimeRead:     true,
	}).Return("", errors.New("some db error details"))

	handler := rest.CreateNote(&service)

	request := web.Request{
		Body: `{
				"text": "Hello World",
				"lifeTimeSeconds": 360000,
				"password": "mySecretPassword",
				"oneTimeRead": true
			}`,
	}

	// when
	gotResp, gotErr := handler(context.TODO(), request)

	// then
	assert.Equal(t, web.Response{
		StatusCode: http.StatusInternalServerError,
		Body:       "Internal server error",
	}, gotResp)
	assert.EqualError(t, gotErr, "create note: some db error details")
}

type mockCreateService struct {
	mock.Mock
}

func (m *mockCreateService) CreateNote(ctx context.Context, plain creating.Note) (noteID string, err error) {
	args := m.Called(plain)
	return args.String(0), args.Error(1)
}
