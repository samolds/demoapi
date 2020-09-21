package server

import (
	"context"
	"net/http"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/zeebo/errs"

	"demoapi/handler"
	he "demoapi/httperror"
)

func validateToken(token string) error {
	if token == "" {
		return errs.New("empty!")
	}

	// TODO(sam): actually validate the token
	return nil
}

func (s *Server) Authenticated(h handler.Handler) handler.Handler {
	return handler.Handler(func(ctx context.Context, w http.ResponseWriter,
		r *http.Request) (interface{}, error) {

		logrus.Debugf("checking that the request is authenticated")

		authorizationHeader := r.Header.Get("authorization")
		if authorizationHeader == "" {
			return nil, he.Unauthenticated.New("no authorization header")
		}

		parts := strings.Fields(authorizationHeader)
		if len(parts) != 2 {
			return nil, he.Unauthenticated.New("bad authorization header")
		}

		err := validateToken(parts[1]) // part[0] should be "Bearer"
		if err != nil {
			return nil, he.Unauthenticated.New("invalid token: %s", err)
		}

		return h(ctx, w, r)
	})
}
