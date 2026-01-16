package admin

import (
	"context"
	"net/http"

	"ride-hail/internal/auth"
	"ride-hail/internal/middleware"
	"ride-hail/internal/shared/core"
)

type AdminApi struct {
	authMiddleware middleware.Middleware
	handler        handler
	server         http.Server
}

func NewAdminApi(authService *auth.AuthService, adminService *AdminService, port string) *AdminApi {
	authMiddleware := middleware.AuthMiddleware(*authService, core.UserRoleAdmin)
	handler := newHandler(*adminService, authService)

	api := &AdminApi{
		authMiddleware: authMiddleware,
		handler:        *handler,
		server: http.Server{
			Addr: ":" + port,
		},
	}
	return api
}

func (a *AdminApi) Start() error {
	mux := http.NewServeMux()
	chain := middleware.CreateMiddlewareChain(middleware.LoggingMiddleware, a.authMiddleware)

	mux.Handle("POST /sign_up", middleware.LoggingMiddleware(a.handler.signUp))
	mux.Handle("POST /login", middleware.LoggingMiddleware(a.handler.login))

	mux.Handle("GET /admin/overview", chain(a.handler.overview))
	mux.Handle("GET /admin/rides/active", chain(a.handler.active))
	a.server.Handler = mux
	return a.server.ListenAndServe()
}

func (a *AdminApi) StopServer(ctx context.Context) error {
	return a.server.Shutdown(ctx)
}
