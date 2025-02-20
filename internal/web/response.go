package web

import "net/http"

type Response struct {
	status int
	json   any
}

func NewResponse(options ...func(response *Response)) Response {
	response := Response{
		status: http.StatusNoContent,
		json:   nil,
	}

	for _, option := range options {
		option(&response)
	}

	return response
}

func Status(status int) func(r *Response) {
	return func(r *Response) {
		r.status = status
	}
}

func Json(body any) func(r *Response) {
	return func(r *Response) {
		r.json = body
	}
}
