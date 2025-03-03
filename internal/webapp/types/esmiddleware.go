package types

type EsMiddleware func(handler EsHandler) EsHandler

func WrapMiddleware(mw []EsMiddleware, handler EsHandler) EsHandler {
	for i := len(mw) - 1; i >= 0; i-- {
		h := mw[i]
		if h != nil {
			handler = h(handler)
		}
	}

	return handler
}
