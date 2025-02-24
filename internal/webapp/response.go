package webapp

import "net/http"

type Response struct {
	status  int
	headers map[string]string
	json    any
}

func NewResponse(options ...func(response *Response)) Response {
	response := Response{
		status:  http.StatusNoContent,
		headers: make(map[string]string),
		json:    nil,
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

func Header(key, value string) func(r *Response) {
	return func(r *Response) {
		r.headers[key] = value
	}
}

func Json(body any) func(r *Response) {
	return func(r *Response) {
		r.json = body
	}
}
