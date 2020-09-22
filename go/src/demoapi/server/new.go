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

// NewHTTPServer constructs a new http.Server to listen for connections and
// serve responses as defined by the Server's ServeHTTP defined above.
func NewHTTPServer(configs *config.Configs,
	metricMiddleware h.MiddlewareWrapper) (*http.Server, error) {

	// TODO(sam): pass through database configs
	db, err := database.Connect(configs.DBURL, nil)
	if err != nil {
		return nil, err
	}

	var apiHandler http.Handler
	apiHandler = New(db, configs)

	if metricMiddleware != nil {
		apiHandler = metricMiddleware(apiHandler)
	}

	return &http.Server{
		Addr:         configs.APIAddress,
		WriteTimeout: configs.WriteTimeout,
		ReadTimeout:  configs.ReadTimeout,
		IdleTimeout:  configs.IdleTimeout,
		Handler:      apiHandler,
	}, nil
}
