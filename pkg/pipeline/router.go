package pipeline

import "strings"

type Router struct {
	handlers map[string]Handler
}

func NewRouter() *Router {
	return &Router{handlers: make(map[string]Handler)}
}

func (r *Router) Register(handler Handler) {
	contentType := handler.ContentType()
	contentType = strings.ToLower(contentType)
	r.handlers[contentType] = handler
}

func (r *Router) Route(contentType string) (Handler, bool) {
	contentType = strings.ToLower(contentType)
	handler, ok := r.handlers[contentType]
	return handler, ok
}
