package middleware

import "github.com/ilia-tolliu/serverless-event-store/internal/webapp/types"

type EsMiddleware func(handler types.EsHandler) types.EsHandler

func Wrap(mw []EsMiddleware, handler types.EsHandler) types.EsHandler {
	for i := len(mw) - 1; i >= 0; i-- {
		h := mw[i]
		if h != nil {
			handler = h(handler)
		}
	}

	return handler
}
