package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMiddlewareChainSimple(t *testing.T) {
	mw := MiddlewareChain()
	h := mw.JSON(handler0)

	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "/", nil))

	assert.Equal(t, `"0"`, w.Body.String())
}

func TestMiddlewareChainDefault(t *testing.T) {
	mw := MiddlewareChain(middlewareNothing, middlewarePre5)
	h := mw.JSON(handler2)

	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "/", nil))

	assert.Equal(t, `pre5"2"`, w.Body.String())
}

func TestMiddlewareChainAppend(t *testing.T) {
	mw := MiddlewareChain(middlewareSandwich9, middlewarePre5,
		middlewarePre5, middlewareNothing)
	mw1 := mw.Append(middlewarePre5)
	mw2 := mw.Append(middlewareSandwich9)

	h1 := mw1.JSON(handler3)
	h2 := mw2.JSON(handler3)

	w1 := httptest.NewRecorder()
	w2 := httptest.NewRecorder()

	h1.ServeHTTP(w1, httptest.NewRequest("GET", "/", nil))
	h2.ServeHTTP(w2, httptest.NewRequest("GET", "/", nil))

	assert.Equal(t, `pre9pre5pre5pre5post9"3"`, w1.Body.String())
	assert.Equal(t, `pre9pre5pre5pre9post9post9"3"`, w2.Body.String())
}

func handler0(context.Context, http.ResponseWriter, *http.Request) (interface{},
	error) {
	return "0", nil
}

func handler1(context.Context, http.ResponseWriter, *http.Request) (interface{},
	error) {
	return "1", nil
}

func handler2(context.Context, http.ResponseWriter, *http.Request) (interface{},
	error) {
	return "2", nil
}

func handler3(context.Context, http.ResponseWriter, *http.Request) (interface{},
	error) {
	return "3", nil
}

func middlewareNothing(h Handler) Handler {
	return Handler(func(ctx context.Context, w http.ResponseWriter,
		r *http.Request) (interface{}, error) {
		return h(ctx, w, r)
	})
}

func middlewarePre5(h Handler) Handler {
	return Handler(func(ctx context.Context, w http.ResponseWriter,
		r *http.Request) (interface{}, error) {
		_, err := w.Write([]byte("pre5"))
		if err != nil {
			return nil, err
		}
		return h(ctx, w, r)
	})
}

func middlewareSandwich9(h Handler) Handler {
	return Handler(func(ctx context.Context, w http.ResponseWriter,
		r *http.Request) (interface{}, error) {
		_, err := w.Write([]byte("pre9"))
		if err != nil {
			return nil, err
		}
		b, mainErr := h(ctx, w, r)
		_, err = w.Write([]byte("post9"))
		if err != nil {
			return nil, err
		}
		return b, mainErr
	})
}
