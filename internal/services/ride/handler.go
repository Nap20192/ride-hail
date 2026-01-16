package ride

import (
	"encoding/json"
	"io"
	"net/http"

	"ride-hail/internal/auth"
	"ride-hail/internal/middleware"
	"ride-hail/internal/shared/core"
	"ride-hail/pkg/server"
	"ride-hail/pkg/uuid"
)

type handler struct {
	service *RideService
	manager *server.Manager
	auth    *auth.AuthService
}

func newHandler(service *RideService, manager *server.Manager, auth *auth.AuthService) *handler {
	return &handler{
		service: service,
		manager: manager,
		auth:    auth,
	}
}

func (h handler) create(w http.ResponseWriter, r *http.Request) {
	var inputCreateRide CreateRideRequest
	data, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	if err := json.Unmarshal(data, &inputCreateRide); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	// Extract passenger ID from JWT claims using context helper
	passengerID, err := middleware.GetUserIDFromContext(r.Context())
	if err != nil {
		http.Error(w, "unauthorized: invalid user context", http.StatusUnauthorized)
		return
	}

	// Set passenger ID from authenticated user
	inputCreateRide.PassengerID = passengerID

	if err := inputCreateRide.Validate(); err != nil {
		http.Error(w, "invalid input: "+err.Error(), http.StatusBadRequest)
		return
	}

	ride, err := h.service.CreateRide(r.Context(), inputCreateRide)
	if err != nil {
		http.Error(w, "failed to create ride: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	bytes, _ := json.Marshal(ride)

	w.Write(bytes)
}

func (h handler) cancel(w http.ResponseWriter, r *http.Request) {
	// Extract ride ID from URL path
	rideIDStr := r.PathValue("id")
	if rideIDStr == "" {
		http.Error(w, "ride ID is required", http.StatusBadRequest)
		return
	}

	rideID, err := uuid.FromString(rideIDStr)
	if err != nil {
		http.Error(w, "invalid ride ID format", http.StatusBadRequest)
		return
	}

	// Extract passenger ID from JWT claims using context helper
	passengerID, err := middleware.GetUserIDFromContext(r.Context())
	if err != nil {
		http.Error(w, "unauthorized: invalid user context", http.StatusUnauthorized)
		return
	}

	// Read and parse request body
	var cancelReq CancelRideRequest
	data, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if err := json.Unmarshal(data, &cancelReq); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	// Validate request
	if err := cancelReq.Validate(); err != nil {
		http.Error(w, "invalid input: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Cancel the ride
	response, err := h.service.CancelRide(r.Context(), rideID, passengerID, cancelReq.Reason)
	if err != nil {
		// Check for specific error types
		if err.Error() == "ride not found: no rows in result set" || err.Error() == "ride not found: sql: no rows in result set" {
			http.Error(w, "ride not found", http.StatusNotFound)
			return
		}
		if err.Error() == "unauthorized: ride does not belong to passenger" {
			http.Error(w, "unauthorized: you can only cancel your own rides", http.StatusForbidden)
			return
		}
		if err.Error() == "cannot cancel completed ride" {
			http.Error(w, "cannot cancel completed ride", http.StatusBadRequest)
			return
		}
		if err.Error() == "ride is already cancelled" {
			http.Error(w, "ride is already cancelled", http.StatusBadRequest)
			return
		}
		http.Error(w, "failed to cancel ride: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	bytes, _ := json.Marshal(response)
	w.Write(bytes)
}

func (h handler) websocket(w http.ResponseWriter, r *http.Request) {
	h.manager.ServeWS(w, r)
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

	token, err := h.auth.SignUp(r.Context(), input, core.UserRolePassenger)
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
