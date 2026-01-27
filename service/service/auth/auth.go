package auth

import (
	"context"
	"encoding/base64"
	"log/slog"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// protectedServicePaths defines the gRPC service paths that require authentication
var protectedServicePaths = []string{
	"/feature.Feature/",
	"/workload.Workload/",
}

// requiresAuthentication checks if a method requires authentication
func requiresAuthentication(fullMethod string) bool {
	for _, path := range protectedServicePaths {
		if strings.Contains(fullMethod, path) {
			return true
		}
	}
	return false
}

// validateCredentials extracts and validates credentials from the context metadata
func validateCredentials(ctx context.Context, fullMethod, username, password string) error {
	// Extract metadata from context
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		slog.WarnContext(ctx, "Missing metadata in request", "method", fullMethod)
		return status.Error(codes.Unauthenticated, "missing metadata")
	}

	// Check for authorization header
	authHeaders := md.Get("authorization")
	if len(authHeaders) == 0 {
		slog.WarnContext(ctx, "Missing authorization header", "method", fullMethod)
		return status.Error(codes.Unauthenticated, "missing authorization header")
	}

	// Parse Basic Auth header
	auth := authHeaders[0]
	if !strings.HasPrefix(auth, "Basic ") {
		slog.WarnContext(ctx, "Invalid authorization header format", "method", fullMethod)
		return status.Error(codes.Unauthenticated, "invalid authorization header")
	}

	// Decode base64 credentials
	payload, err := base64.StdEncoding.DecodeString(auth[6:])
	if err != nil {
		slog.WarnContext(ctx, "Failed to decode authorization header", "method", fullMethod, "error", err)
		return status.Error(codes.Unauthenticated, "invalid authorization header")
	}

	// Split username:password
	pair := strings.SplitN(string(payload), ":", 2)
	if len(pair) != 2 {
		slog.WarnContext(ctx, "Invalid credentials format", "method", fullMethod)
		return status.Error(codes.Unauthenticated, "invalid credentials format")
	}

	// Validate credentials
	if pair[0] != username || pair[1] != password {
		slog.WarnContext(ctx, "Invalid credentials", "method", fullMethod)
		return status.Error(codes.Unauthenticated, "invalid credentials")
	}

	// Credentials are valid
	slog.DebugContext(ctx, "Authentication successful", "method", fullMethod, "username", pair[0])
	return nil
}

// SelectiveInterceptor creates a gRPC unary server interceptor that only applies
// authentication to specific services (Feature and Workload)
func SelectiveInterceptor(enabled bool, username, password string) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		// If authentication is not enabled, allow all requests
		if !enabled {
			return handler(ctx, req)
		}

		// Only authenticate protected services (Feature and Workload)
		// Allow Health and Meta services to pass through without authentication
		if !requiresAuthentication(info.FullMethod) {
			return handler(ctx, req)
		}

		// Validate credentials
		if err := validateCredentials(ctx, info.FullMethod, username, password); err != nil {
			return nil, err
		}

		// Credentials are valid, proceed with the request
		return handler(ctx, req)
	}
}

// SelectiveStreamInterceptor creates a gRPC stream server interceptor that only applies
// authentication to specific services (Feature and Workload)
func SelectiveStreamInterceptor(enabled bool, username, password string) grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		// If authentication is not enabled, allow all requests
		if !enabled {
			return handler(srv, ss)
		}

		// Only authenticate protected services (Feature and Workload)
		// Allow Health and Meta services to pass through without authentication
		if !requiresAuthentication(info.FullMethod) {
			return handler(srv, ss)
		}

		// Validate credentials
		if err := validateCredentials(ss.Context(), info.FullMethod, username, password); err != nil {
			return err
		}

		// Credentials are valid, proceed with the request
		return handler(srv, ss)
	}
}
