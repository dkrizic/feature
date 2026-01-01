package httpclient_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/dkrizic/feature/cli/telemetry/httpclient"
)

func TestClient_Get(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET request, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test response"))
	}))
	defer server.Close()

	// Test with package-level function
	ctx := context.Background()
	resp, err := httpclient.Get(ctx, server.URL)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestClient_Post(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		contentType := r.Header.Get("Content-Type")
		if contentType != "application/json" {
			t.Errorf("Expected Content-Type application/json, got %s", contentType)
		}
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	// Test with package-level function
	ctx := context.Background()
	resp, err := httpclient.Post(ctx, server.URL, "application/json", nil)
	if err != nil {
		t.Fatalf("Post failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", resp.StatusCode)
	}
}

func TestNewClient(t *testing.T) {
	// Test creating a new client with nil transport
	client := httpclient.New(nil)
	if client == nil {
		t.Fatal("Expected non-nil client")
	}

	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Test that the client works
	ctx := context.Background()
	resp, err := client.Get(ctx, server.URL)
	if err != nil {
		t.Fatalf("Client.Get failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestDefaultClient(t *testing.T) {
	client := httpclient.DefaultClient()
	if client == nil {
		t.Fatal("Expected non-nil default client")
	}
}

func TestClient_PostForm(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		contentType := r.Header.Get("Content-Type")
		if contentType != "application/x-www-form-urlencoded" {
			t.Errorf("Expected Content-Type application/x-www-form-urlencoded, got %s", contentType)
		}

		// Read and verify the body
		err := r.ParseForm()
		if err != nil {
			t.Errorf("Failed to parse form: %v", err)
		}

		if r.Form.Get("key") != "value" {
			t.Errorf("Expected form field 'key' to be 'value', got '%s'", r.Form.Get("key"))
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Test with package-level function
	ctx := context.Background()
	formData := url.Values{}
	formData.Set("key", "value")

	resp, err := httpclient.PostForm(ctx, server.URL, formData)
	if err != nil {
		t.Fatalf("PostForm failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}
