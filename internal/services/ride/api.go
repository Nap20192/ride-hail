package ride

import (
	"context"
	"net/http"

	"ride-hail/internal/auth"
	"ride-hail/internal/middleware"
	"ride-hail/internal/shared/core"
	"ride-hail/pkg/server"
)

type RideApi struct {
	websocketManager *server.Manager
	authMiddleware   middleware.Middleware
	handler          *handler
	server           *http.Server
}

func NewRideApi(authService *auth.AuthService, ride *RideService, port string) *RideApi {
	wsManager := server.NewManager()

	authMiddleware := middleware.AuthMiddleware(*authService,core.UserRolePassenger)
	handler := newHandler(ride, wsManager, authService)

	api := &RideApi{
		websocketManager: wsManager,
		authMiddleware:   authMiddleware,
		handler:          handler,
		server: &http.Server{
			Addr: ":" + port,
		},
	}
	return api
}

func (r *RideApi) Start() error {
	mux := http.NewServeMux()

	chain := middleware.CreateMiddlewareChain(middleware.LoggingMiddleware, r.authMiddleware)

	mux.Handle("POST /sign_up", middleware.LoggingMiddleware(r.handler.signUp))
	mux.Handle("POST /login", middleware.LoggingMiddleware(r.handler.login))

	mux.Handle("POST /rides", chain(r.handler.create))
	mux.Handle("POST /rides/{id}/cancel", chain(r.handler.cancel))

	mux.Handle("GET /ws/passengers/{id}", chain(r.handler.websocket))

	r.server.Handler = mux
	return r.server.ListenAndServe()
}

func (r *RideApi) StopServer(ctx context.Context) error {
	return r.server.Shutdown(ctx)
}

func (r *RideApi) GetWebsocketManager() *server.Manager {
	return r.websocketManager
}
