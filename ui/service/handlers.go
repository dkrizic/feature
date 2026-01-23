package service

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"sort"
	"sync"

	featurev1 "github.com/dkrizic/feature/ui/repository/feature/v1"
	workloadv1 "github.com/dkrizic/feature/ui/repository/workload/v1"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"google.golang.org/protobuf/types/known/emptypb"
)

// Feature represents a feature flag with a key and value.
type Feature struct {
	Key   string
	Value string
}

// sessionStore holds active sessions in memory with thread-safe access
var (
	sessionStore = make(map[string]bool)
	sessionMutex sync.RWMutex
)

// generateSessionID creates a random session ID
func generateSessionID() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// authMiddleware checks if the user is authenticated
func (s *Server) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !s.authEnabled {
			next.ServeHTTP(w, r)
			return
		}

		// Check for session cookie
		cookie, err := r.Cookie("session")
		if err == nil {
			sessionMutex.RLock()
			authenticated := sessionStore[cookie.Value]
			sessionMutex.RUnlock()
			
			if authenticated {
				next.ServeHTTP(w, r)
				return
			}
		}

		// Redirect to login page
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	})
}

// registerHandlers registers all HTTP handlers on the provided mux.
func (s *Server) registerHandlers(mux *http.ServeMux) {
	prefix := s.subpath
	
	// Public routes
	mux.HandleFunc("GET "+prefix+"/login", s.handleLoginPage)
	mux.HandleFunc("POST "+prefix+"/login", s.handleLogin)
	mux.HandleFunc("POST "+prefix+"/logout", s.handleLogout)
	mux.HandleFunc("GET "+prefix+"/health", s.handleHealth)

	// Protected routes (with auth middleware)
	mux.HandleFunc("GET "+prefix+"/", s.authMiddleware(otelhttp.NewHandler(http.HandlerFunc(s.handleIndex), "handleIndex")).ServeHTTP)
	mux.HandleFunc("GET "+prefix+"/features/list", s.authMiddleware(otelhttp.NewHandler(http.HandlerFunc(s.handleFeaturesList), "handleFeaturesList")).ServeHTTP)
	mux.HandleFunc("POST "+prefix+"/features/create", s.authMiddleware(otelhttp.NewHandler(http.HandlerFunc(s.handleFeatureCreate), "handleFeatureCreate")).ServeHTTP)
	mux.HandleFunc("POST "+prefix+"/features/update", s.authMiddleware(otelhttp.NewHandler(http.HandlerFunc(s.handleFeatureUpdate), "handleFeatureUpdate")).ServeHTTP)
	mux.HandleFunc("POST "+prefix+"/features/delete", s.authMiddleware(otelhttp.NewHandler(http.HandlerFunc(s.handleFeatureDelete), "handleFeatureDelete")).ServeHTTP)
}

// handleLoginPage renders the login page
func (s *Server) handleLoginPage(w http.ResponseWriter, r *http.Request) {
	ctx, span := otel.Tracer("ui/service").Start(r.Context(), "handleLoginPage")
	defer span.End()

	// If auth is disabled, redirect to home
	if !s.authEnabled {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// If already logged in, redirect to home
	cookie, err := r.Cookie("session")
	if err == nil {
		sessionMutex.RLock()
		authenticated := sessionStore[cookie.Value]
		sessionMutex.RUnlock()
		
		if authenticated {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
	}

	data := struct {
		UIVersion string
		Error     string
	}{
		UIVersion: s.uiVersion,
		Error:     "",
	}

	if err := s.templates.ExecuteTemplate(w, "login.gohtml", data); err != nil {
		slog.ErrorContext(ctx, "Failed to render login template", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		span.SetStatus(codes.Error, err.Error())
		return
	}
}

// handleLogin processes login form submission
func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	ctx, span := otel.Tracer("ui/service").Start(r.Context(), "handleLogin")
	defer span.End()

	// Parse form data
	if err := r.ParseForm(); err != nil {
		slog.ErrorContext(ctx, "Failed to parse form", "error", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		span.SetStatus(codes.Error, err.Error())
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")

	// Validate credentials using constant time comparison to prevent timing attacks
	usernameMatch := subtle.ConstantTimeCompare([]byte(username), []byte(s.authUsername)) == 1
	passwordMatch := subtle.ConstantTimeCompare([]byte(password), []byte(s.authPassword)) == 1
	
	if usernameMatch && passwordMatch {
		// Create session
		sessionID, err := generateSessionID()
		if err != nil {
			slog.ErrorContext(ctx, "Failed to generate session ID", "error", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			span.SetStatus(codes.Error, err.Error())
			return
		}
		
		sessionMutex.Lock()
		sessionStore[sessionID] = true
		sessionMutex.Unlock()

		// Determine if connection is secure (HTTPS)
		// Check X-Forwarded-Proto header for proxied connections
		secure := r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https"

		// Set session cookie (HttpOnly for security)
		http.SetCookie(w, &http.Cookie{
			Name:     "session",
			Value:    sessionID,
			Path:     "/",
			HttpOnly: true,
			Secure:   secure,
			SameSite: http.SameSiteStrictMode,
			MaxAge:   s.authSessionTimeout,
		})

		slog.InfoContext(ctx, "User logged in", "username", username, "sessionTimeout", s.authSessionTimeout)
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// Invalid credentials
	slog.WarnContext(ctx, "Invalid login attempt", "username", username)

	data := struct {
		UIVersion string
		Error     string
	}{
		UIVersion: s.uiVersion,
		Error:     "Invalid username or password",
	}

	if err := s.templates.ExecuteTemplate(w, "login.gohtml", data); err != nil {
		slog.ErrorContext(ctx, "Failed to render login template", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		span.SetStatus(codes.Error, err.Error())
		return
	}
}

// handleLogout logs out the user by clearing their session
func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	ctx, span := otel.Tracer("ui/service").Start(r.Context(), "handleLogout")
	defer span.End()

	// Get session cookie
	cookie, err := r.Cookie("session")
	if err == nil {
		// Remove session from store
		sessionMutex.Lock()
		delete(sessionStore, cookie.Value)
		sessionMutex.Unlock()
		slog.InfoContext(ctx, "User logged out")
	}

	// Clear cookie
	// Determine if connection is secure (HTTPS)
	secure := r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https"
	
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		MaxAge:   -1, // Delete cookie
	})

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

// handleIndex renders the full HTML page with HTMX.
func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	ctx, span := otel.Tracer("ui/service").Start(r.Context(), "handleIndex")
	defer span.End()

	data := struct {
		UIVersion      string
		BackendVersion string
		Subpath        string
		Authenticated  bool
	}{
		UIVersion:      s.uiVersion,
		BackendVersion: s.backendVersion,
		Subpath:        s.subpath,
		Authenticated:  s.authEnabled,
	}

	if err := s.templates.ExecuteTemplate(w, "index.gohtml", data); err != nil {
		slog.ErrorContext(ctx, "Failed to render index template", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		span.SetStatus(codes.Error, err.Error())
		return
	}
}

// handleFeaturesList calls the gRPC backend to get all features and renders the partial.
func (s *Server) handleFeaturesList(w http.ResponseWriter, r *http.Request) {
	ctx, span := otel.Tracer("ui/service").Start(r.Context(), "handleFeaturesList")
	defer span.End()

	slog.InfoContext(ctx, "Fetching all features from backend")

	// Call the gRPC backend
	stream, err := s.featureClient.GetAll(ctx, &emptypb.Empty{})
	if err != nil {
		slog.ErrorContext(ctx, "Failed to call GetAll", "error", err)
		http.Error(w, "Failed to fetch features", http.StatusInternalServerError)
		span.SetStatus(codes.Error, err.Error())
		return
	}

	// Collect all features from the stream
	var features []Feature

	for {
		kv, err := stream.Recv()
		if err != nil {
			// Check if we've reached the end of the stream
			if err == io.EOF {
				break
			}
			slog.ErrorContext(ctx, "Failed to receive feature", "error", err)
			http.Error(w, "Failed to fetch features", http.StatusInternalServerError)
			span.SetStatus(codes.Error, err.Error())
			return
		}

		features = append(features, Feature{
			Key:   kv.Key,
			Value: kv.Value,
		})
	}

	// Sort features alphabetically by key
	sort.Slice(features, func(i, j int) bool {
		return features[i].Key < features[j].Key
	})

	data := struct {
		Features []Feature
		Subpath  string
	}{
		Features: features,
		Subpath:  s.subpath,
	}

	if err := s.templates.ExecuteTemplate(w, "features_list.gohtml", data); err != nil {
		slog.ErrorContext(ctx, "Failed to render features_list template", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		span.SetStatus(codes.Error, err.Error())
		return
	}
}

// handleFeatureCreate creates a new feature flag and re-renders the list.
func (s *Server) handleFeatureCreate(w http.ResponseWriter, r *http.Request) {
	ctx, span := otel.Tracer("ui/service").Start(r.Context(), "handleFeatureCreate")
	defer span.End()

	slog.InfoContext(ctx, "Creating feature")

	// Parse form data
	if err := r.ParseForm(); err != nil {
		slog.ErrorContext(ctx, "Failed to parse form", "error", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		span.SetStatus(codes.Error, err.Error())
		return
	}

	key := r.FormValue("key")
	if key == "" {
		slog.ErrorContext(ctx, "Missing key parameter")
		http.Error(w, "Missing key parameter", http.StatusBadRequest)
		span.SetStatus(codes.Error, "Missing key parameter")
		return
	}

	value := r.FormValue("value")

	// Call the gRPC backend to set (upsert)
	_, err := s.featureClient.Set(ctx, &featurev1.KeyValue{Key: key, Value: value})
	if err != nil {
		slog.ErrorContext(ctx, "Failed to create feature", "key", key, "error", err)
		http.Error(w, "Failed to create feature", http.StatusInternalServerError)
		span.SetStatus(codes.Error, err.Error())
		return
	}

	slog.InfoContext(ctx, "Feature created", "key", key, "value", value)

	// Re-render the feature list by calling the list handler
	s.handleFeaturesList(w, r)
}

// handleFeatureUpdate updates an existing feature flag and re-renders the list.
func (s *Server) handleFeatureUpdate(w http.ResponseWriter, r *http.Request) {
	ctx, span := otel.Tracer("ui/service").Start(r.Context(), "handleFeatureUpdate")
	defer span.End()

	// Parse form data
	if err := r.ParseForm(); err != nil {
		slog.ErrorContext(ctx, "Failed to parse form", "error", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		span.SetStatus(codes.Error, err.Error())
		return
	}

	key := r.FormValue("key")
	if key == "" {
		slog.ErrorContext(ctx, "Missing key parameter")
		http.Error(w, "Missing key parameter", http.StatusBadRequest)
		span.SetStatus(codes.Error, "Missing key parameter")
		return
	}

	value := r.FormValue("value")

	// Call the gRPC backend to set (update)
	_, err := s.featureClient.Set(ctx, &featurev1.KeyValue{Key: key, Value: value})
	if err != nil {
		slog.ErrorContext(ctx, "Failed to update feature", "key", key, "error", err)
		http.Error(w, "Failed to update feature", http.StatusInternalServerError)
		span.SetStatus(codes.Error, err.Error())
		return
	}

	slog.InfoContext(ctx, "Feature updated", "key", key, "value", value)

	// Re-render the feature list by calling the list handler
	s.handleFeaturesList(w, r)
}

// handleFeatureDelete deletes a feature and re-renders the list.
func (s *Server) handleFeatureDelete(w http.ResponseWriter, r *http.Request) {
	ctx, span := otel.Tracer("ui/service").Start(r.Context(), "handleFeatureDelete")
	defer span.End()

	// Parse form data
	if err := r.ParseForm(); err != nil {
		slog.ErrorContext(ctx, "Failed to parse form", "error", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		span.SetStatus(codes.Error, err.Error())
		return
	}

	key := r.FormValue("key")
	if key == "" {
		slog.ErrorContext(ctx, "Missing key parameter")
		http.Error(w, "Missing key parameter", http.StatusBadRequest)
		span.SetStatus(codes.Error, "Missing key parameter")
		return
	}

	// Call the gRPC backend to delete
	_, err := s.featureClient.Delete(ctx, &featurev1.Key{Name: key})
	if err != nil {
		slog.ErrorContext(ctx, "Failed to delete feature", "key", key, "error", err)
		http.Error(w, "Failed to delete feature", http.StatusInternalServerError)
		span.SetStatus(codes.Error, err.Error())
		return
	}

	slog.InfoContext(ctx, "Feature deleted", "key", key)

	// Re-render the feature list by calling the list handler
	s.handleFeaturesList(w, r)
}

// handleHealth is a simple health check endpoint.
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// handleWorkloadRestart handles workload restart requests
func (s *Server) handleWorkloadRestart(w http.ResponseWriter, r *http.Request) {
	ctx, span := otel.Tracer("ui/service").Start(r.Context(), "handleWorkloadRestart")
	defer span.End()

	slog.InfoContext(ctx, "Handling workload restart request")

	// Parse form data
	if err := r.ParseForm(); err != nil {
		slog.ErrorContext(ctx, "Failed to parse form", "error", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		span.SetStatus(codes.Error, err.Error())
		return
	}

	workloadType := r.FormValue("type")
	workloadName := r.FormValue("name")
	namespace := r.FormValue("namespace")

	// Validate inputs
	if workloadType == "" {
		slog.ErrorContext(ctx, "Missing workload type parameter")
		http.Error(w, "Missing workload type parameter", http.StatusBadRequest)
		span.SetStatus(codes.Error, "Missing workload type parameter")
		return
	}

	if workloadName == "" {
		slog.ErrorContext(ctx, "Missing workload name parameter")
		http.Error(w, "Missing workload name parameter", http.StatusBadRequest)
		span.SetStatus(codes.Error, "Missing workload name parameter")
		return
	}

	// Map string type to protobuf enum
	var protoType workloadv1.WorkloadType
	switch workloadType {
	case "deployment":
		protoType = workloadv1.WorkloadType_WORKLOAD_TYPE_DEPLOYMENT
	case "statefulset":
		protoType = workloadv1.WorkloadType_WORKLOAD_TYPE_STATEFULSET
	case "daemonset":
		protoType = workloadv1.WorkloadType_WORKLOAD_TYPE_DAEMONSET
	default:
		slog.ErrorContext(ctx, "Invalid workload type", "type", workloadType)
		http.Error(w, fmt.Sprintf("Invalid workload type: %s", workloadType), http.StatusBadRequest)
		span.SetStatus(codes.Error, "Invalid workload type")
		return
	}

	// Call the gRPC backend
	resp, err := s.workloadClient.RestartWorkload(ctx, &workloadv1.RestartRequest{
		Type:      protoType,
		Name:      workloadName,
		Namespace: namespace,
	})
	if err != nil {
		slog.ErrorContext(ctx, "Failed to restart workload", "type", workloadType, "name", workloadName, "error", err)
		http.Error(w, fmt.Sprintf("Failed to restart workload: %v", err), http.StatusInternalServerError)
		span.SetStatus(codes.Error, err.Error())
		return
	}

	if !resp.Success {
		slog.WarnContext(ctx, "Workload restart unsuccessful", "message", resp.Message)
		http.Error(w, resp.Message, http.StatusBadRequest)
		span.SetStatus(codes.Error, resp.Message)
		return
	}

	slog.InfoContext(ctx, "Workload restarted successfully", "type", workloadType, "name", workloadName, "namespace", namespace)

	// Return success message
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf(`<div class="success-message">%s</div>`, resp.Message)))
}
