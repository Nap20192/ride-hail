package api

import (
	"encoding/json"
	"io"
	"net/http"
	"ride-hail/internal/auth"
	"ride-hail/internal/middleware"
	"ride-hail/internal/services/driver"
	"ride-hail/internal/shared/core"
	"ride-hail/pkg/server"
)

type handler struct {
	service   *driver.DriverService
	auth      *auth.AuthService
	wsManager *server.Manager
}

func newHandler(service *driver.DriverService, authService *auth.AuthService, wsManager *server.Manager) *handler {
	return &handler{
		service:   service,
		auth:      authService,
		wsManager: wsManager,
	}
}

func (h handler) signUp(w http.ResponseWriter, r *http.Request) {
	var input auth.RegisterInput
	data, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	if err := json.Unmarshal(data, &input); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	token, err := h.auth.SignUp(r.Context(), input, core.UserRoleDriver)

	if err != nil {
		http.Error(w, "failed to register user: "+err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(token))
}

func (h handler) login(w http.ResponseWriter, r *http.Request) {

	var input auth.LoginInput

	data, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	if err := json.Unmarshal(data, &input); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	token, err := h.auth.LogIn(r.Context(), input)
	if err != nil {
		http.Error(w, "failed to login: "+err.Error(), http.StatusUnauthorized)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(token))
}

func (h *handler) online(w http.ResponseWriter, r *http.Request) {
	// claims, err := middleware.GetClaimsFromContext(r.Context())
	// if err != nil {
	// 	http.Error(w, "unauthorized: invalid user context", http.StatusUnauthorized)
	// 	return
	// }

}

func (h *handler) offline(w http.ResponseWriter, r *http.Request) {
}

func (h *handler) location(w http.ResponseWriter, r *http.Request) {

}

func (h *handler) start(w http.ResponseWriter, r *http.Request) {

}

func (h *handler) complete(w http.ResponseWriter, r *http.Request) {

}

func (h *handler) websocket(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(middleware.UserContextKey).(auth.JWTClaims)
	if !ok {
		http.Error(w, "unauthorized: invalid user context", http.StatusUnauthorized)
		return
	}

	driverID := r.PathValue("id")
	if driverID == "" {
		http.Error(w, "driver_id is required", http.StatusBadRequest)
		return
	}
	if claims.UserID.String() != driverID {
		http.Error(w, "forbidden: cannot open websocket for another driver", http.StatusForbidden)
		return
	}

	h.wsManager.ServeWS(w, r)
}

func writeJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if payload == nil {
		return
	}
	bytes, _ := json.Marshal(payload)
	_, _ = w.Write(bytes)
}
