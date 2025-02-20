package web

import (
	"context"
	"fmt"
	"net/http"
)

func MwLogger(handler Handler) Handler {
	h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		fmt.Println("Request:", r.Method, r.URL.Path)

		err := handler(ctx, w, r)

		fmt.Println("Response send")

		return err
	}

	return h
}
