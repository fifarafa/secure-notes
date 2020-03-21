package rest

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/projects/secure-notes/internal/creating"
	"github.com/projects/secure-notes/internal/web"
)

// CreateNote returns a handler for /POST note request
func CreateNote(s *creating.Service) web.Handler {
	return func(ctx context.Context, req web.Request) (web.Response, error) {
		var newNote creating.Note
		if err := json.Unmarshal([]byte(req.Body), &newNote); err != nil {
			return web.Response{
				StatusCode: http.StatusBadRequest,
			}, err
		}

		noteID, err := s.CreateNote(ctx, newNote)
		if err != nil {
			return web.InternalServerError(), fmt.Errorf("create response: %w", err)
		}

		resp, err := createResponse(noteID)
		if err != nil {
			return web.InternalServerError(), fmt.Errorf("create response: %w", err)
		}

		return resp, nil
	}
}

func createResponse(noteID string) (web.Response, error) {
	type ResponseId struct {
		ID string `json:"id"`
	}

	responseBytes, err := json.Marshal(&ResponseId{ID: noteID})
	if err != nil {
		return web.Response{}, fmt.Errorf("json marshal response: %w", err)
	}

	resp := web.Response{
		StatusCode: http.StatusCreated,
		Body:       string(responseBytes),
	}
	return resp, nil
}
