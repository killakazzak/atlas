// Package http exposes the Inventory HTTP API under /api/v1/servers.
// The standard library is imported as nethttp to avoid a name collision.
package http

import (
	"errors"
	"log/slog"
	nethttp "net/http"
	"strings"
	"time"

	"atlas/internal/auth"
	"atlas/internal/http/middleware"
	"atlas/internal/httpx"
	"atlas/internal/inventory"
)

// Handler serves inventory HTTP endpoints using the Inventory Service.
type Handler struct {
	service inventory.Service
	tokens  auth.TokenService
	logger  *slog.Logger
}

// NewHandler constructs an inventory HTTP Handler.
func NewHandler(service inventory.Service, tokens auth.TokenService, logger *slog.Logger) *Handler {
	return &Handler{service: service, tokens: tokens, logger: logger}
}

// Register mounts inventory routes on mux with role-based access control.
//
//	GET    /api/v1/servers       — Viewer+
//	GET    /api/v1/servers/{id}  — Viewer+
//	POST   /api/v1/servers       — Operator+
//	PUT    /api/v1/servers/{id}  — Operator+
//	DELETE /api/v1/servers/{id}  — Administrator only
func (h *Handler) Register(mux *nethttp.ServeMux) {
	requireAuth := auth.RequireAuth(h.tokens, middleware.RequestIDFromContext)
	viewerPlus := auth.RequireRoles(middleware.RequestIDFromContext,
		auth.RoleAdministrator, auth.RoleOperator, auth.RoleViewer)
	operatorPlus := auth.RequireRoles(middleware.RequestIDFromContext,
		auth.RoleAdministrator, auth.RoleOperator)
	adminOnly := auth.RequireRole(auth.RoleAdministrator, middleware.RequestIDFromContext)

	mux.Handle("GET /api/v1/servers", requireAuth(viewerPlus(nethttp.HandlerFunc(h.listServers))))
	mux.Handle("GET /api/v1/servers/{id}", requireAuth(viewerPlus(nethttp.HandlerFunc(h.getServer))))
	mux.Handle("POST /api/v1/servers", requireAuth(operatorPlus(nethttp.HandlerFunc(h.createServer))))
	mux.Handle("PUT /api/v1/servers/{id}", requireAuth(operatorPlus(nethttp.HandlerFunc(h.updateServer))))
	mux.Handle("DELETE /api/v1/servers/{id}", requireAuth(adminOnly(nethttp.HandlerFunc(h.deleteServer))))
}

type serverResponse struct {
	ID              string    `json:"id"`
	Name            string    `json:"name"`
	Hostname        string    `json:"hostname"`
	IP              string    `json:"ip"`
	OperatingSystem string    `json:"operatingSystem"`
	Status          string    `json:"status"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

type createServerRequest struct {
	Name            string `json:"name"`
	Hostname        string `json:"hostname"`
	IP              string `json:"ip"`
	OperatingSystem string `json:"operatingSystem"`
	Status          string `json:"status"`
}

func (h *Handler) listServers(w nethttp.ResponseWriter, r *nethttp.Request) {
	servers, err := h.service.ListServers(r.Context())
	if err != nil {
		reqID := middleware.RequestIDFromContext(r.Context())
		h.logger.Error("listServers failed", "error", err, "request_id", reqID)
		httpx.WriteError(w, nethttp.StatusInternalServerError, "internal_error", "internal server error", reqID)
		return
	}

	out := make([]serverResponse, 0, len(servers))
	for i := range servers {
		out = append(out, toServerResponse(&servers[i]))
	}
	httpx.WriteJSON(w, nethttp.StatusOK, out)
}

func (h *Handler) getServer(w nethttp.ResponseWriter, r *nethttp.Request) {
	id := strings.TrimSpace(r.PathValue("id"))
	if id == "" {
		httpx.WriteError(w, nethttp.StatusBadRequest, "validation_error", "id is required", middleware.RequestIDFromContext(r.Context()))
		return
	}

	server, err := h.service.GetServer(r.Context(), id)
	if err != nil {
		h.writeServiceError(w, r, err)
		return
	}
	httpx.WriteJSON(w, nethttp.StatusOK, toServerResponse(server))
}

func (h *Handler) createServer(w nethttp.ResponseWriter, r *nethttp.Request) {
	defer func() { _ = r.Body.Close() }()

	var req createServerRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, nethttp.StatusBadRequest, "bad_request", err.Error(), middleware.RequestIDFromContext(r.Context()))
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	req.Hostname = strings.TrimSpace(req.Hostname)
	req.IP = strings.TrimSpace(req.IP)
	req.OperatingSystem = strings.TrimSpace(req.OperatingSystem)
	req.Status = strings.TrimSpace(req.Status)

	if req.Name == "" {
		httpx.WriteError(w, nethttp.StatusBadRequest, "validation_error", "name is required", middleware.RequestIDFromContext(r.Context()))
		return
	}
	if req.Hostname == "" {
		httpx.WriteError(w, nethttp.StatusBadRequest, "validation_error", "hostname is required", middleware.RequestIDFromContext(r.Context()))
		return
	}

	status := inventory.ServerStatusUnknown
	if req.Status != "" {
		status = inventory.ServerStatus(req.Status)
	}

	server := &inventory.Server{
		Name:            req.Name,
		Hostname:        req.Hostname,
		IP:              req.IP,
		OperatingSystem: req.OperatingSystem,
		Status:          status,
	}

	if err := h.service.RegisterServer(r.Context(), server); err != nil {
		h.writeServiceError(w, r, err)
		return
	}

	httpx.WriteJSON(w, nethttp.StatusCreated, toServerResponse(server))
}

func (h *Handler) updateServer(w nethttp.ResponseWriter, r *nethttp.Request) {
	defer func() { _ = r.Body.Close() }()

	id := strings.TrimSpace(r.PathValue("id"))
	if id == "" {
		httpx.WriteError(w, nethttp.StatusBadRequest, "validation_error", "id is required", middleware.RequestIDFromContext(r.Context()))
		return
	}

	existing, err := h.service.GetServer(r.Context(), id)
	if err != nil {
		h.writeServiceError(w, r, err)
		return
	}

	var req createServerRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, nethttp.StatusBadRequest, "bad_request", err.Error(), middleware.RequestIDFromContext(r.Context()))
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	req.Hostname = strings.TrimSpace(req.Hostname)
	req.IP = strings.TrimSpace(req.IP)
	req.OperatingSystem = strings.TrimSpace(req.OperatingSystem)
	req.Status = strings.TrimSpace(req.Status)

	if req.Name == "" {
		httpx.WriteError(w, nethttp.StatusBadRequest, "validation_error", "name is required", middleware.RequestIDFromContext(r.Context()))
		return
	}
	if req.Hostname == "" {
		httpx.WriteError(w, nethttp.StatusBadRequest, "validation_error", "hostname is required", middleware.RequestIDFromContext(r.Context()))
		return
	}

	status := inventory.ServerStatusUnknown
	if req.Status != "" {
		status = inventory.ServerStatus(req.Status)
	}

	existing.Name = req.Name
	existing.Hostname = req.Hostname
	existing.IP = req.IP
	existing.OperatingSystem = req.OperatingSystem
	existing.Status = status

	if err := h.service.UpdateServer(r.Context(), existing); err != nil {
		h.writeServiceError(w, r, err)
		return
	}

	httpx.WriteJSON(w, nethttp.StatusOK, toServerResponse(existing))
}

func (h *Handler) deleteServer(w nethttp.ResponseWriter, r *nethttp.Request) {
	id := strings.TrimSpace(r.PathValue("id"))
	if id == "" {
		httpx.WriteError(w, nethttp.StatusBadRequest, "validation_error", "id is required", middleware.RequestIDFromContext(r.Context()))
		return
	}

	if err := h.service.DeleteServer(r.Context(), id); err != nil {
		h.writeServiceError(w, r, err)
		return
	}
	w.WriteHeader(nethttp.StatusNoContent)
}

func (h *Handler) writeServiceError(w nethttp.ResponseWriter, r *nethttp.Request, err error) {
	reqID := middleware.RequestIDFromContext(r.Context())
	switch {
	case errors.Is(err, inventory.ErrNotFound):
		httpx.WriteError(w, nethttp.StatusNotFound, "not_found", "server not found", reqID)
	case errors.Is(err, inventory.ErrInvalidServer):
		httpx.WriteError(w, nethttp.StatusBadRequest, "validation_error", err.Error(), reqID)
	default:
		h.logger.Error("unexpected service error", "error", err, "request_id", reqID)
		httpx.WriteError(w, nethttp.StatusInternalServerError, "internal_error", "internal server error", reqID)
	}
}

func toServerResponse(server *inventory.Server) serverResponse {
	return serverResponse{
		ID:              server.ID,
		Name:            server.Name,
		Hostname:        server.Hostname,
		IP:              server.IP,
		OperatingSystem: server.OperatingSystem,
		Status:          string(server.Status),
		CreatedAt:       server.CreatedAt,
		UpdatedAt:       server.UpdatedAt,
	}
}
