package handler

import (
	"context"
	"net/http"
)

// TODO(sam): instead of an `interface{}`, create a "Serializable" interface
// that is returned by this Handler interface... just something that is a
// "Writer"? It just needs to be able to output a []byte, like Marshalers

// Handler is an http.Handler interface with a more expressive function
// signature that ensures that all responses, including unexpected errors, are
// returned in a consistent JSON format.
type Handler func(context.Context, http.ResponseWriter, *http.Request) (
	interface{}, error)

type HandlerFunc func(Handler) Handler

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	b, err := h(r.Context(), w, r)
	jsonResponse(w, b, err)
}
