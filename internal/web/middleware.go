package web

import (
	"context"

	"github.com/aws/aws-lambda-go/events"
	"go.uber.org/zap"
)

type Handler func(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error)

type Middleware struct {
	Logger *zap.SugaredLogger
}

func (m *Middleware) WrapWithCorsAndLogging(h Handler) func(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return func(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
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

func addCorsHeaders(original events.APIGatewayProxyResponse) events.APIGatewayProxyResponse {
	return events.APIGatewayProxyResponse{
		StatusCode: original.StatusCode,
		Headers: map[string]string{
			"Access-Control-Allow-Origin":      "*",
			"Access-Control-Allow-Credentials": "true",
		},
		Body:            original.Body,
		IsBase64Encoded: original.IsBase64Encoded,
	}
}
