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

		// Only authenticate Feature and Workload services
		// Allow Health and Meta services to pass through without authentication
		method := info.FullMethod
		requiresAuth := strings.Contains(method, "/feature.Feature/") || 
			strings.Contains(method, "/workload.Workload/")
		
		if !requiresAuth {
			return handler(ctx, req)
		}

		// Extract metadata from context
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			slog.WarnContext(ctx, "Missing metadata in request", "method", info.FullMethod)
			return nil, status.Error(codes.Unauthenticated, "missing metadata")
		}

		// Check for authorization header
		authHeaders := md.Get("authorization")
		if len(authHeaders) == 0 {
			slog.WarnContext(ctx, "Missing authorization header", "method", info.FullMethod)
			return nil, status.Error(codes.Unauthenticated, "missing authorization header")
		}

		// Parse Basic Auth header
		auth := authHeaders[0]
		if !strings.HasPrefix(auth, "Basic ") {
			slog.WarnContext(ctx, "Invalid authorization header format", "method", info.FullMethod)
			return nil, status.Error(codes.Unauthenticated, "invalid authorization header")
		}

		// Decode base64 credentials
		payload, err := base64.StdEncoding.DecodeString(auth[6:])
		if err != nil {
			slog.WarnContext(ctx, "Failed to decode authorization header", "method", info.FullMethod, "error", err)
			return nil, status.Error(codes.Unauthenticated, "invalid authorization header")
		}

		// Split username:password
		pair := strings.SplitN(string(payload), ":", 2)
		if len(pair) != 2 {
			slog.WarnContext(ctx, "Invalid credentials format", "method", info.FullMethod)
			return nil, status.Error(codes.Unauthenticated, "invalid credentials format")
		}

		// Validate credentials
		if pair[0] != username || pair[1] != password {
			slog.WarnContext(ctx, "Invalid credentials", "method", info.FullMethod, "username", pair[0])
			return nil, status.Error(codes.Unauthenticated, "invalid credentials")
		}

		// Credentials are valid, proceed with the request
		slog.DebugContext(ctx, "Authentication successful", "method", info.FullMethod, "username", pair[0])
		return handler(ctx, req)
	}
}
