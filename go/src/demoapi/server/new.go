package server

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"

	"demoapi/config"
	"demoapi/database"
	h "demoapi/handler"
)

type Server struct {
	DB     *database.Database
	router http.Handler
	Config *config.Configs
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

func New(db *database.Database, configs *config.Configs) *Server {
	s := &Server{DB: db, Config: configs}
	s.router = router(s)
	return s
}

func router(s *Server) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	mw := h.MiddlewareChain()
	r.Method("GET", "/", mw.JSON(s.Health))

	if !s.Config.InsecureRequestsMode {
		mw = mw.Append(s.Authenticated) // add auth middleware
	}

	apiRoutes := chi.NewRouter()
	apiRoutes.Method("GET", "/users", mw.JSON(s.PagedUsers))
	apiRoutes.Method("GET", "/users/{userID}", mw.JSON(s.GetUser))
	apiRoutes.Method("POST", "/users", mw.JSON(s.CreateUser))
	apiRoutes.Method("DELETE", "/users/{userID}", mw.JSON(s.DeleteUser))
	apiRoutes.Method("PUT", "/users/{userID}", mw.JSON(s.UpdateUser))

	apiRoutes.Method("GET", "/groups", mw.JSON(s.PagedGroups))
	apiRoutes.Method("GET", "/groups/{groupName}", mw.JSON(s.GetMemberships))
	apiRoutes.Method("POST", "/groups", mw.JSON(s.CreateGroup))
	apiRoutes.Method("PUT", "/groups/{groupName}", mw.JSON(s.UpdateMembership))
	apiRoutes.Method("DELETE", "/groups/{groupName}", mw.JSON(s.DeleteGroup))

	// TODO(sam): it might be better if this lived under an "api" route
	r.Mount("/", apiRoutes)
	return r
}
