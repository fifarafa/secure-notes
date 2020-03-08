package web

import "github.com/aws/aws-lambda-go/events"

type (
	Request  = events.APIGatewayProxyRequest
	Response = events.APIGatewayProxyResponse
)
