package service

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"sort"

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

// registerHandlers registers all HTTP handlers on the provided mux.
func (s *Server) registerHandlers(mux *http.ServeMux) {

	prefix := s.subpath
	mux.HandleFunc("GET "+prefix+"/", otelhttp.NewHandler(http.HandlerFunc(s.handleIndex), "handleIndex").ServeHTTP)
	mux.HandleFunc("GET "+prefix+"/features/list", otelhttp.NewHandler(http.HandlerFunc(s.handleFeaturesList), "handleFeaturesList").ServeHTTP)
	mux.HandleFunc("POST "+prefix+"/features/create", otelhttp.NewHandler(http.HandlerFunc(s.handleFeatureCreate), "handleFeatureCreate").ServeHTTP)
	mux.HandleFunc("POST "+prefix+"/features/update", otelhttp.NewHandler(http.HandlerFunc(s.handleFeatureUpdate), "handleFeatureUpdate").ServeHTTP)
	mux.HandleFunc("POST "+prefix+"/features/delete", otelhttp.NewHandler(http.HandlerFunc(s.handleFeatureDelete), "handleFeatureDelete").ServeHTTP)
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
	}{
		UIVersion:      s.uiVersion,
		BackendVersion: s.backendVersion,
		Subpath:        s.subpath,
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
