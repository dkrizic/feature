package service

import (
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net/http"

	"github.com/dkrizic/feature/ui/constant"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
)

// generateSessionID generates a random session ID
func generateSessionID() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// isAuthenticated checks if the request has a valid session cookie
func (s *Server) isAuthenticated(r *http.Request) bool {
	if !s.authEnabled {
		return true
	}

	cookie, err := r.Cookie(constant.SessionCookieName)
	if err != nil {
		return false
	}

	s.sessionsMutex.RLock()
	defer s.sessionsMutex.RUnlock()
	return s.authenticatedSessions[cookie.Value]
}

// requireAuth is middleware that checks authentication and redirects to login if needed
func (s *Server) requireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !s.authEnabled {
			next(w, r)
			return
		}

		if !s.isAuthenticated(r) {
			// Redirect to login page
			http.Redirect(w, r, s.subpath+"/login", http.StatusSeeOther)
			return
		}

		next(w, r)
	}
}

// handleLogin renders the login page or processes login
func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	ctx, span := otel.Tracer("ui/service").Start(r.Context(), "handleLogin")
	defer span.End()

	// If auth is not enabled, redirect to home
	if !s.authEnabled {
		http.Redirect(w, r, s.subpath+"/", http.StatusSeeOther)
		return
	}

	// If already authenticated, redirect to home
	if s.isAuthenticated(r) {
		http.Redirect(w, r, s.subpath+"/", http.StatusSeeOther)
		return
	}

	if r.Method == http.MethodPost {
		// Process login
		if err := r.ParseForm(); err != nil {
			slog.ErrorContext(ctx, "Failed to parse login form", "error", err)
			http.Error(w, "Bad Request", http.StatusBadRequest)
			span.SetStatus(codes.Error, err.Error())
			return
		}

		username := r.FormValue("username")
		password := r.FormValue("password")

		if username == s.authUsername && password == s.authPassword {
			// Generate session ID
			sessionID, err := generateSessionID()
			if err != nil {
				slog.ErrorContext(ctx, "Failed to generate session ID", "error", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				span.SetStatus(codes.Error, err.Error())
				return
			}

			// Store session
			s.sessionsMutex.Lock()
			s.authenticatedSessions[sessionID] = true
			s.sessionsMutex.Unlock()

			// Set cookie
			// Note: Secure flag is not set to support both HTTP and HTTPS deployments.
			// In production, use HTTPS and consider adding Secure: true via reverse proxy or ingress.
			http.SetCookie(w, &http.Cookie{
				Name:     constant.SessionCookieName,
				Value:    sessionID,
				Path:     s.subpath + "/",
				HttpOnly: true,
				SameSite: http.SameSiteStrictMode,
				MaxAge:   86400, // 24 hours
			})

			slog.InfoContext(ctx, "User logged in successfully", "username", username)

			// Redirect to home
			http.Redirect(w, r, s.subpath+"/", http.StatusSeeOther)
			return
		}

		// Invalid credentials
		slog.WarnContext(ctx, "Invalid login attempt", "username", username)
		data := struct {
			Subpath string
			Error   string
		}{
			Subpath: s.subpath,
			Error:   "Invalid username or password",
		}

		if err := s.templates.ExecuteTemplate(w, "login.gohtml", data); err != nil {
			slog.ErrorContext(ctx, "Failed to render login template", "error", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			span.SetStatus(codes.Error, err.Error())
			return
		}
		return
	}

	// Render login page
	data := struct {
		Subpath string
		Error   string
	}{
		Subpath: s.subpath,
		Error:   "",
	}

	if err := s.templates.ExecuteTemplate(w, "login.gohtml", data); err != nil {
		slog.ErrorContext(ctx, "Failed to render login template", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		span.SetStatus(codes.Error, err.Error())
		return
	}
}

// handleLogout logs out the user
func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	ctx, span := otel.Tracer("ui/service").Start(r.Context(), "handleLogout")
	defer span.End()

	cookie, err := r.Cookie(constant.SessionCookieName)
	if err == nil {
		// Remove session
		s.sessionsMutex.Lock()
		delete(s.authenticatedSessions, cookie.Value)
		s.sessionsMutex.Unlock()
	}

	// Clear cookie
	// Note: Secure flag is not set to support both HTTP and HTTPS deployments.
	http.SetCookie(w, &http.Cookie{
		Name:     constant.SessionCookieName,
		Value:    "",
		Path:     s.subpath + "/",
		HttpOnly: true,
		MaxAge:   -1,
	})

	slog.InfoContext(ctx, "User logged out")

	// Redirect to login
	http.Redirect(w, r, s.subpath+"/login", http.StatusSeeOther)
}
