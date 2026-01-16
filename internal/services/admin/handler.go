package admin

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"ride-hail/internal/auth"
	"ride-hail/internal/middleware"
	"ride-hail/internal/shared/core"
)

type handler struct {
	service AdminService
	auth    *auth.AuthService
}

func newHandler(service AdminService, auth *auth.AuthService) *handler {
	return &handler{
		service: service,
		auth:    auth,
	}
}

func (h handler) overview(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()

	// Extract and verify admin role using context helper
	role, err := middleware.GetUserRoleFromContext(ctx)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if role != core.UserRoleAdmin.String() {
		userID, _ := middleware.GetUserIDFromContext(ctx)
		slog.Warn("unauthorized admin access attempt",
			slog.String("user_id", userID.String()),
			slog.String("role", role))
		http.Error(w, "Forbidden: Admin access required", http.StatusForbidden)
		return
	}

	// Audit log: Admin accessed overview
	adminID, _ := middleware.GetUserIDFromContext(ctx)
	slog.Info("admin accessed system overview",
		slog.String("admin_id", adminID.String()),
		slog.String("action", "view_overview"))

	// Get system metrics
	metrics, err := h.service.GetSystemMetrics(ctx)
	if err != nil {
		adminID, _ := middleware.GetUserIDFromContext(ctx)
		slog.Error("Failed to get system metrics",
			slog.String("admin_id", adminID.String()),
			slog.String("error", err.Error()))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Get driver distribution
	driverDist, err := h.service.GetDriverDistribution(ctx)
	if err != nil {
		adminID, _ := middleware.GetUserIDFromContext(ctx)
		slog.Error("Failed to get driver distribution",
			slog.String("admin_id", adminID.String()),
			slog.String("error", err.Error()))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	response := OverviewResponse{
		Timestamp:          time.Now(),
		Metrics:            metrics,
		DriverDistribution: driverDist,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		slog.Error("Failed to encode response", slog.String("error", err.Error()))
	}
}

func (h handler) active(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()

	// Extract and verify admin role using context helper
	role, err := middleware.GetUserRoleFromContext(ctx)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if role != core.UserRoleAdmin.String() {
		userID, _ := middleware.GetUserIDFromContext(ctx)
		slog.Warn("unauthorized admin access attempt to active rides",
			slog.String("user_id", userID.String()),
			slog.String("role", role))
		http.Error(w, "Forbidden: Admin access required", http.StatusForbidden)
		return
	}

	// Parse pagination parameters
	pageStr := r.URL.Query().Get("page")
	pageSizeStr := r.URL.Query().Get("page_size")

	page := 1
	pageSize := 20

	if pageStr != "" {
		if p, err := parsePositiveInt(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if pageSizeStr != "" {
		if ps, err := parsePositiveInt(pageSizeStr); err == nil && ps > 0 && ps <= 100 {
			pageSize = ps
		}
	}

	// Audit log: Admin accessed active rides
	adminID, _ := middleware.GetUserIDFromContext(ctx)
	slog.Info("admin accessed active rides list",
		slog.String("admin_id", adminID.String()),
		slog.String("action", "view_active_rides"),
		slog.Int("page", page),
		slog.Int("page_size", pageSize))

	// Get active rides
	rides, totalCount, err := h.service.GetActiveRides(ctx, page, pageSize)
	if err != nil {
		slog.Error("Failed to get active rides",
			slog.String("admin_id", adminID.String()),
			slog.Int("page", page),
			slog.Int("page_size", pageSize),
			slog.String("error", err.Error()))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	response := ActiveRidesResponse{
		Rides:      rides,
		TotalCount: totalCount,
		Page:       page,
		PageSize:   pageSize,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		slog.Error("Failed to encode response", slog.String("error", err.Error()))
	}
}

func parsePositiveInt(s string) (int, error) {
	var n int
	if _, err := fmt.Sscanf(s, "%d", &n); err != nil {
		return 0, err
	}
	if n < 0 {
		return 0, fmt.Errorf("negative number")
	}
	return n, nil
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
