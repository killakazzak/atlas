// Package http exposes the Auth HTTP API under /api/v1/auth.
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
)

// Handler serves auth HTTP endpoints.
type Handler struct {
	service auth.Service
	tokens  auth.TokenService
	ttl     time.Duration
	logger  *slog.Logger
}

// NewHandler constructs an auth HTTP Handler.
func NewHandler(service auth.Service, tokens auth.TokenService, ttl time.Duration, logger *slog.Logger) *Handler {
	return &Handler{service: service, tokens: tokens, ttl: ttl, logger: logger}
}

// Register mounts auth routes on mux.
// /auth/me is wrapped with RequireAuth middleware so it requires a valid token.
func (h *Handler) Register(mux *nethttp.ServeMux) {
	requireAuth := auth.RequireAuth(h.tokens, middleware.RequestIDFromContext)

	mux.HandleFunc("POST /api/v1/auth/login", h.login)
	mux.HandleFunc("POST /api/v1/auth/logout", h.logout)
	mux.Handle("GET /api/v1/auth/me", requireAuth(nethttp.HandlerFunc(h.me)))
}

// loginRequest is the body expected by POST /api/v1/auth/login.
type loginRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

// loginResponse is returned on successful login.
type loginResponse struct {
	AccessToken string `json:"accessToken"`
	ExpiresIn   int64  `json:"expiresIn"`
	TokenType   string `json:"tokenType"`
}

// meResponse is returned by GET /api/v1/auth/me.
type meResponse struct {
	ID    string    `json:"id"`
	Login string    `json:"login"`
	Email string    `json:"email"`
	Role  auth.Role `json:"role"`
}

func (h *Handler) login(w nethttp.ResponseWriter, r *nethttp.Request) {
	defer func() { _ = r.Body.Close() }()

	reqID := middleware.RequestIDFromContext(r.Context())

	var req loginRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, nethttp.StatusBadRequest, "bad_request", err.Error(), reqID)
		return
	}

	req.Login = strings.TrimSpace(req.Login)
	if req.Login == "" {
		httpx.WriteError(w, nethttp.StatusBadRequest, "validation_error", "login is required", reqID)
		return
	}
	if req.Password == "" {
		httpx.WriteError(w, nethttp.StatusBadRequest, "validation_error", "password is required", reqID)
		return
	}

	user, err := h.service.Authenticate(r.Context(), req.Login, req.Password)
	if err != nil {
		if errors.Is(err, auth.ErrInvalidCredentials) {
			httpx.WriteError(w, nethttp.StatusUnauthorized, "unauthorized", "invalid login or password", reqID)
			return
		}
		h.logger.Error("authenticate failed", "error", err, "request_id", reqID)
		httpx.WriteError(w, nethttp.StatusInternalServerError, "internal_error", "internal server error", reqID)
		return
	}

	token, err := h.tokens.Generate(user)
	if err != nil {
		h.logger.Error("generate token failed", "error", err, "request_id", reqID)
		httpx.WriteError(w, nethttp.StatusInternalServerError, "internal_error", "internal server error", reqID)
		return
	}

	httpx.WriteJSON(w, nethttp.StatusOK, loginResponse{
		AccessToken: token,
		ExpiresIn:   int64(h.ttl.Seconds()),
		TokenType:   "Bearer",
	})
}

// logout is a stateless stub. JWT revocation requires a token store (not yet
// implemented), so the client is expected to discard the token locally.
func (h *Handler) logout(w nethttp.ResponseWriter, _ *nethttp.Request) {
	w.WriteHeader(nethttp.StatusNoContent)
}

func (h *Handler) me(w nethttp.ResponseWriter, r *nethttp.Request) {
	claims := auth.ClaimsFromContext(r.Context())
	if claims == nil {
		// Should not happen: JWTAuth middleware always sets claims before calling this.
		httpx.WriteError(w, nethttp.StatusUnauthorized, "unauthorized", "missing token claims", middleware.RequestIDFromContext(r.Context()))
		return
	}

	user, err := h.service.GetUser(r.Context(), claims.UserID)
	if err != nil {
		reqID := middleware.RequestIDFromContext(r.Context())
		if errors.Is(err, auth.ErrNotFound) {
			httpx.WriteError(w, nethttp.StatusUnauthorized, "unauthorized", "user not found", reqID)
			return
		}
		h.logger.Error("me: get user failed", "error", err, "request_id", reqID)
		httpx.WriteError(w, nethttp.StatusInternalServerError, "internal_error", "internal server error", reqID)
		return
	}

	httpx.WriteJSON(w, nethttp.StatusOK, meResponse{
		ID:    user.ID,
		Login: user.Username,
		Email: user.Email,
		Role:  user.Role,
	})
}
