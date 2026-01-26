package command

import (
	"context"
	"encoding/base64"

	"github.com/dkrizic/feature/cli/constant"
	feature "github.com/dkrizic/feature/cli/repository/feature/v1"
	workload "github.com/dkrizic/feature/cli/repository/workload/v1"
	"github.com/urfave/cli/v3"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// basicAuthCreds implements credentials.PerRPCCredentials for Basic Auth
type basicAuthCreds struct {
	username string
	password string
}

func (c *basicAuthCreds) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	auth := c.username + ":" + c.password
	enc := base64.StdEncoding.EncodeToString([]byte(auth))
	return map[string]string{
		"authorization": "Basic " + enc,
	}, nil
}

func (c *basicAuthCreds) RequireTransportSecurity() bool {
	return false
}

func FeatureClient(cmd *cli.Command) (feature.FeatureClient, error) {
	endpoint := cmd.String(constant.Endpoint)
	username := cmd.String(constant.Username)
	password := cmd.String(constant.Password)

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
	}

	// Add authentication if credentials are provided
	if username != "" && password != "" {
		opts = append(opts, grpc.WithPerRPCCredentials(&basicAuthCreds{
			username: username,
			password: password,
		}))
	}

	gc, err := grpc.NewClient(endpoint, opts...)
	if err != nil {
		return nil, err
	}

	fc := feature.NewFeatureClient(gc)
	return fc, nil
}

func WorkloadClient(cmd *cli.Command) (workload.WorkloadClient, error) {
	endpoint := cmd.String(constant.Endpoint)
	username := cmd.String(constant.Username)
	password := cmd.String(constant.Password)

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
	}

	// Add authentication if credentials are provided
	if username != "" && password != "" {
		opts = append(opts, grpc.WithPerRPCCredentials(&basicAuthCreds{
			username: username,
			password: password,
		}))
	}

	gc, err := grpc.NewClient(endpoint, opts...)
	if err != nil {
		return nil, err
	}

	wc := workload.NewWorkloadClient(gc)
	return wc, nil
}
