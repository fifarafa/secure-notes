package web

import (
	"context"

	"go.uber.org/zap"
)

type Handler func(ctx context.Context, req Request) (Response, error)

type Middleware struct {
	Logger *zap.SugaredLogger
}

func (m *Middleware) WrapWithCorsAndLogging(h Handler) func(ctx context.Context, req Request) (Response, error) {
	return func(ctx context.Context, req Request) (Response, error) {
		resp, err := h(ctx, req)
		addCorsHeaders(resp)

		if err != nil {
			m.Logger.Errorw("failed to handle request successfully",
				"error", err,
				"request", req,
				"response", resp)
		}

		return resp, nil
	}
}

func addCorsHeaders(original Response) Response {
	return Response{
		StatusCode: original.StatusCode,
		Headers: map[string]string{
			"Access-Control-Allow-Origin":      "*",
			"Access-Control-Allow-Credentials": "true",
		},
		Body:            original.Body,
		IsBase64Encoded: original.IsBase64Encoded,
	}
}
