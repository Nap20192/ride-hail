package api

import (
	"context"
	"net/http"

	"ride-hail/internal/auth"
	"ride-hail/internal/middleware"
	"ride-hail/internal/services/driver"
	"ride-hail/internal/shared/core"
	"ride-hail/pkg/server"
)

type DriverApi struct {
	websocketManager *server.Manager
	authMiddleware   middleware.Middleware
	handler          *handler
	server           *http.Server
}

func NewDriverApi(authService *auth.AuthService, driverService *driver.DriverService, port string) *DriverApi {
	wsManager := server.NewManager()
	authMiddleware := middleware.AuthMiddleware(*authService, core.UserRoleDriver)
	handler := newHandler(driverService, authService, wsManager)

	api := &DriverApi{
		websocketManager: wsManager,
		authMiddleware:   authMiddleware,
		handler:          handler,
		server: &http.Server{
			Addr: ":" + port,
		},
	}
	return api
}

func (d *DriverApi) Start() error {
	mux := http.NewServeMux()

	chain := middleware.CreateMiddlewareChain(middleware.LoggingMiddleware, d.authMiddleware)

	mux.Handle("POST /sign_up", middleware.LoggingMiddleware(d.handler.signUp))
	mux.Handle("POST /login", middleware.LoggingMiddleware(d.handler.login))

	mux.Handle("POST /drivers/{driver_id}/online", chain(d.handler.online))
	mux.Handle("POST /drivers/{driver_id}/offline", chain(d.handler.offline))
	mux.Handle("POST /drivers/{driver_id}/location", chain(d.handler.location))
	mux.Handle("POST /drivers/{driver_id}/start", chain(d.handler.start))
	mux.Handle("POST /drivers/{driver_id}/complete", chain(d.handler.complete))
	mux.Handle("GET /ws/drivers/{id}", chain(d.handler.websocket))

	d.server.Handler = mux
	return d.server.ListenAndServe()
}

func (d *DriverApi) StopServer(ctx context.Context) error {
	return d.server.Shutdown(ctx)
}

func (d *DriverApi) GetWebsocketManager() *server.Manager {
	return d.websocketManager
}
