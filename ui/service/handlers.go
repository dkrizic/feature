package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"sort"
	"time"

	featurev1 "github.com/dkrizic/feature/ui/repository/feature/v1"
	metav1 "github.com/dkrizic/feature/ui/repository/meta/v1"
	workloadv1 "github.com/dkrizic/feature/ui/repository/workload/v1"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"google.golang.org/protobuf/types/known/emptypb"
)

// Feature represents a feature flag with a key and value.
type Feature struct {
	Key      string
	Value    string
	Editable bool
}

// registerHandlers registers all HTTP handlers on the provided mux.
func (s *Server) registerHandlers(mux *http.ServeMux) {

	prefix := s.subpath
	
	// Login and logout routes (no auth required)
	mux.HandleFunc("GET "+prefix+"/login", s.handleLogin)
	mux.HandleFunc("POST "+prefix+"/login", s.handleLogin)
	mux.HandleFunc("GET "+prefix+"/logout", s.handleLogout)
	
	// Protected routes (require auth if enabled)
	mux.HandleFunc("GET "+prefix+"/", s.requireAuth(otelhttp.NewHandler(http.HandlerFunc(s.handleIndex), "handleIndex").ServeHTTP))
	mux.HandleFunc("GET "+prefix+"/features/list", s.requireAuth(otelhttp.NewHandler(http.HandlerFunc(s.handleFeaturesList), "handleFeaturesList").ServeHTTP))
	mux.HandleFunc("POST "+prefix+"/features/create", s.requireAuth(otelhttp.NewHandler(http.HandlerFunc(s.handleFeatureCreate), "handleFeatureCreate").ServeHTTP))
	mux.HandleFunc("POST "+prefix+"/features/update", s.requireAuth(otelhttp.NewHandler(http.HandlerFunc(s.handleFeatureUpdate), "handleFeatureUpdate").ServeHTTP))
	mux.HandleFunc("POST "+prefix+"/features/delete", s.requireAuth(otelhttp.NewHandler(http.HandlerFunc(s.handleFeatureDelete), "handleFeatureDelete").ServeHTTP))
	mux.HandleFunc("POST "+prefix+"/restart", s.requireAuth(otelhttp.NewHandler(http.HandlerFunc(s.handleRestart), "handleRestart").ServeHTTP))
	mux.HandleFunc("GET "+prefix+"/version", s.requireAuth(otelhttp.NewHandler(http.HandlerFunc(s.handleVersion), "handleVersion").ServeHTTP))
	
	// Health check (no auth required)
	mux.HandleFunc("GET "+prefix+"/health", s.handleHealth)
}

// handleIndex renders the full HTML page with HTMX.
func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	ctx, span := otel.Tracer("ui/service").Start(r.Context(), "handleIndex")
	defer span.End()

	data := struct {
		UIVersion      string
		BackendVersion string
		Subpath        string
		RestartEnabled bool
		RestartName    string
		RestartType    string
		AuthEnabled    bool
	}{
		UIVersion:      s.uiVersion,
		BackendVersion: s.backendVersion,
		Subpath:        s.subpath,
		RestartEnabled: s.restartEnabled,
		RestartName:    s.restartName,
		RestartType:    s.restartType,
		AuthEnabled:    s.authEnabled,
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
			Key:      kv.Key,
			Value:    kv.Value,
			Editable: kv.Editable,
		})
	}

	// Sort features alphabetically by key
	sort.Slice(features, func(i, j int) bool {
		return features[i].Key < features[j].Key
	})

	// Check if restrictions are active (at least one field is not editable)
	restrictionsActive := false
	for _, f := range features {
		if !f.Editable {
			restrictionsActive = true
			break
		}
	}

	data := struct {
		Features           []Feature
		Subpath            string
		RestrictionsActive bool
	}{
		Features:           features,
		Subpath:            s.subpath,
		RestrictionsActive: restrictionsActive,
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

// handleVersion fetches the current backend version and returns it as JSON.
func (s *Server) handleVersion(w http.ResponseWriter, r *http.Request) {
	ctx, span := otel.Tracer("ui/service").Start(r.Context(), "handleVersion")
	defer span.End()

	slog.InfoContext(ctx, "Fetching backend version")

	// Fetch backend version with a timeout
	const grpcCallTimeout = 5 * time.Second
	metaCtx, cancel := context.WithTimeout(ctx, grpcCallTimeout)
	defer cancel()
	
	metaResp, err := s.metaClient.Meta(metaCtx, &metav1.MetaRequest{})
	if err != nil {
		slog.WarnContext(ctx, "Failed to fetch backend version", "error", err)
		http.Error(w, "Failed to fetch backend version", http.StatusInternalServerError)
		span.SetStatus(codes.Error, err.Error())
		return
	}

	// Return the version as JSON
	response := map[string]string{
		"backendVersion": metaResp.Version,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		slog.ErrorContext(ctx, "Failed to encode version response", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		span.SetStatus(codes.Error, err.Error())
		return
	}
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

// handleRestart handles the simplified restart request using configured values
func (s *Server) handleRestart(w http.ResponseWriter, r *http.Request) {
	ctx, span := otel.Tracer("ui/service").Start(r.Context(), "handleRestart")
	defer span.End()

	slog.InfoContext(ctx, "Handling restart request for configured service")

	// Validate that restart is enabled
	if !s.restartEnabled {
		slog.WarnContext(ctx, "Restart feature is not enabled")
		http.Error(w, "Restart feature is not enabled", http.StatusForbidden)
		span.SetStatus(codes.Error, "Restart feature is not enabled")
		return
	}

	// Call the gRPC backend
	resp, err := s.workloadClient.Restart(ctx, &workloadv1.SimpleRestartRequest{})
	if err != nil {
		slog.ErrorContext(ctx, "Failed to restart service", "error", err)
		http.Error(w, fmt.Sprintf("Failed to restart service: %v", err), http.StatusInternalServerError)
		span.SetStatus(codes.Error, err.Error())
		return
	}

	if !resp.Success {
		slog.WarnContext(ctx, "Service restart unsuccessful", "message", resp.Message)
		http.Error(w, resp.Message, http.StatusBadRequest)
		span.SetStatus(codes.Error, resp.Message)
		return
	}

	slog.InfoContext(ctx, "Service restarted successfully")

	// Return success message
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf(`<div class="success-message">%s</div>`, resp.Message)))
}
