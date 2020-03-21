package rest

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/projects/secure-notes/internal/creating"
	"github.com/projects/secure-notes/internal/getting"
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

		resp, err := createNoteResponse(noteID)
		if err != nil {
			return web.InternalServerError(), fmt.Errorf("create response: %w", err)
		}

		return resp, nil
	}
}

func createNoteResponse(noteID string) (web.Response, error) {
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

// GetNote returns a handler for /GET note request
func GetNote(s *getting.Service) web.Handler {
	return func(ctx context.Context, req web.Request) (web.Response, error) {
		noteID := req.PathParameters["id"]
		plainPwd := req.Headers["password"]

		note, err := s.Get(ctx, noteID, plainPwd)
		if err != nil {
			switch err {

			case getting.ErrNotFound:
				return web.Response{
					StatusCode: http.StatusNotFound,
				}, fmt.Errorf("get note from db: %w", err)

			case getting.ErrNotAuthorized:
				return web.Response{
					StatusCode: http.StatusUnauthorized,
				}, fmt.Errorf("wrong password")

			default:
				return web.InternalServerError(), fmt.Errorf("get note from db: %w", err)
			}
		}

		resp, err := getNoteResponse(note, err)
		if err != nil {
			return web.InternalServerError(), fmt.Errorf("create response: %w", err)
		}

		return resp, nil
	}
}

func getNoteResponse(n getting.Note, err error) (web.Response, error) {
	noteBytes, err := json.Marshal(n)
	if err != nil {
		return web.Response{}, fmt.Errorf("json marshal response: %w", err)
	}

	resp := web.Response{
		StatusCode: http.StatusOK,
		Body:       string(noteBytes),
	}
	return resp, nil
}
