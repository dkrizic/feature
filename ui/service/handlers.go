package service

import (
	"log/slog"
	"net/http"

	featurev1 "github.com/dkrizic/feature/ui/repository/feature/v1"
	"google.golang.org/protobuf/types/known/emptypb"
)

// registerHandlers registers all HTTP handlers on the provided mux.
func (s *Server) registerHandlers(mux *http.ServeMux) {
	mux.HandleFunc("GET /", s.handleIndex)
	mux.HandleFunc("GET /features/list", s.handleFeaturesList)
	mux.HandleFunc("POST /features/delete", s.handleFeatureDelete)
	mux.HandleFunc("GET /healthz", s.handleHealth)
}

// handleIndex renders the full HTML page with HTMX.
func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	data := struct {
		UIVersion      string
		BackendVersion string
	}{
		UIVersion:      s.uiVersion,
		BackendVersion: s.backendVersion,
	}

	if err := s.templates.ExecuteTemplate(w, "index.gohtml", data); err != nil {
		slog.Error("Failed to render index template", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

// handleFeaturesList calls the gRPC backend to get all features and renders the partial.
func (s *Server) handleFeaturesList(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Call the gRPC backend
	stream, err := s.featureClient.GetAll(ctx, &emptypb.Empty{})
	if err != nil {
		slog.Error("Failed to call GetAll", "error", err)
		http.Error(w, "Failed to fetch features", http.StatusInternalServerError)
		return
	}

	// Collect all features from the stream
	var features []struct {
		Key   string
		Value string
	}

	for {
		kv, err := stream.Recv()
		if err != nil {
			// Check if we've reached the end of the stream
			if err.Error() == "EOF" {
				break
			}
			slog.Error("Failed to receive feature", "error", err)
			http.Error(w, "Failed to fetch features", http.StatusInternalServerError)
			return
		}

		features = append(features, struct {
			Key   string
			Value string
		}{
			Key:   kv.Key,
			Value: kv.Value,
		})
	}

	data := struct {
		Features []struct {
			Key   string
			Value string
		}
	}{
		Features: features,
	}

	if err := s.templates.ExecuteTemplate(w, "features_list.gohtml", data); err != nil {
		slog.Error("Failed to render features_list template", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

// handleFeatureDelete deletes a feature and re-renders the list.
func (s *Server) handleFeatureDelete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse form data
	if err := r.ParseForm(); err != nil {
		slog.Error("Failed to parse form", "error", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	key := r.FormValue("key")
	if key == "" {
		slog.Error("Missing key parameter")
		http.Error(w, "Missing key parameter", http.StatusBadRequest)
		return
	}

	// Call the gRPC backend to delete
	_, err := s.featureClient.Delete(ctx, &featurev1.Key{Name: key})
	if err != nil {
		slog.Error("Failed to delete feature", "key", key, "error", err)
		http.Error(w, "Failed to delete feature", http.StatusInternalServerError)
		return
	}

	slog.Info("Feature deleted", "key", key)

	// Re-render the feature list by calling the list handler
	s.handleFeaturesList(w, r)
}

// handleHealth is a simple health check endpoint.
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
