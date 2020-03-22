package web

import (
	"net/http"
)

func InternalServerError() Response {
	return Response{
		StatusCode: http.StatusInternalServerError,
		Body:       "Internal server error",
	}
}
