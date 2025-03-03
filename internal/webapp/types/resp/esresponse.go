package resp

import "net/http"

type EsResponse struct {
	status  int
	headers map[string]string
	json    any
}

func New(options ...func(response *EsResponse)) EsResponse {
	response := EsResponse{
		status:  http.StatusNoContent,
		headers: make(map[string]string),
		json:    nil,
	}

	for _, option := range options {
		option(&response)
	}

	return response
}

func WithStatus(status int) func(r *EsResponse) {
	return func(r *EsResponse) {
		r.status = status
	}
}

func WithHeader(key, value string) func(r *EsResponse) {
	return func(r *EsResponse) {
		r.headers[key] = value
	}
}

func WithJson(body any) func(r *EsResponse) {
	return func(r *EsResponse) {
		r.json = body
	}
}

func (r *EsResponse) Status() int {
	return r.status
}

func (r *EsResponse) Headers() map[string]string {
	return r.headers
}

func (r *EsResponse) Json() any {
	return r.json
}
