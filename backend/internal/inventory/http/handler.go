// Package http exposes the Inventory HTTP API under /api/v1/servers.
// The standard library is imported as nethttp to avoid a name collision.
package http

import (
	"encoding/json"
	"errors"
	nethttp "net/http"
	"strings"
	"time"

	"atlas/internal/inventory"
)

// Handler serves inventory HTTP endpoints using the Inventory Service.
type Handler struct {
	service inventory.Service
}

// NewHandler constructs an inventory HTTP Handler.
func NewHandler(service inventory.Service) *Handler {
	return &Handler{service: service}
}

// Register mounts inventory routes on mux.
func (h *Handler) Register(mux *nethttp.ServeMux) {
	mux.HandleFunc("GET /api/v1/servers", h.listServers)
	mux.HandleFunc("GET /api/v1/servers/{id}", h.getServer)
	mux.HandleFunc("POST /api/v1/servers", h.createServer)
	mux.HandleFunc("DELETE /api/v1/servers/{id}", h.deleteServer)
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

type errorResponse struct {
	Error string `json:"error"`
}

func (h *Handler) listServers(w nethttp.ResponseWriter, r *nethttp.Request) {
	servers, err := h.service.ListServers(r.Context())
	if err != nil {
		writeError(w, nethttp.StatusInternalServerError, "failed to list servers")
		return
	}

	out := make([]serverResponse, 0, len(servers))
	for i := range servers {
		out = append(out, toServerResponse(&servers[i]))
	}
	writeJSON(w, nethttp.StatusOK, out)
}

func (h *Handler) getServer(w nethttp.ResponseWriter, r *nethttp.Request) {
	id := r.PathValue("id")
	if strings.TrimSpace(id) == "" {
		writeError(w, nethttp.StatusBadRequest, "id is required")
		return
	}

	server, err := h.service.GetServer(r.Context(), id)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, nethttp.StatusOK, toServerResponse(server))
}

func (h *Handler) createServer(w nethttp.ResponseWriter, r *nethttp.Request) {
	defer r.Body.Close()

	var req createServerRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		writeError(w, nethttp.StatusBadRequest, "invalid JSON body")
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	req.Hostname = strings.TrimSpace(req.Hostname)
	req.IP = strings.TrimSpace(req.IP)
	req.OperatingSystem = strings.TrimSpace(req.OperatingSystem)
	req.Status = strings.TrimSpace(req.Status)

	if req.Name == "" {
		writeError(w, nethttp.StatusBadRequest, "name is required")
		return
	}
	if req.Hostname == "" {
		writeError(w, nethttp.StatusBadRequest, "hostname is required")
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
		writeServiceError(w, err)
		return
	}

	writeJSON(w, nethttp.StatusCreated, toServerResponse(server))
}

func (h *Handler) deleteServer(w nethttp.ResponseWriter, r *nethttp.Request) {
	id := r.PathValue("id")
	if strings.TrimSpace(id) == "" {
		writeError(w, nethttp.StatusBadRequest, "id is required")
		return
	}

	if err := h.service.DeleteServer(r.Context(), id); err != nil {
		writeServiceError(w, err)
		return
	}
	w.WriteHeader(nethttp.StatusNoContent)
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

func writeServiceError(w nethttp.ResponseWriter, err error) {
	switch {
	case errors.Is(err, inventory.ErrNotFound):
		writeError(w, nethttp.StatusNotFound, "server not found")
	case errors.Is(err, inventory.ErrInvalidServer):
		writeError(w, nethttp.StatusBadRequest, err.Error())
	default:
		writeError(w, nethttp.StatusInternalServerError, "internal server error")
	}
}

func writeError(w nethttp.ResponseWriter, status int, message string) {
	writeJSON(w, status, errorResponse{Error: message})
}

func writeJSON(w nethttp.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
