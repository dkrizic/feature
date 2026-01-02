package command

import (
	"github.com/dkrizic/feature/cli/constant"
	feature "github.com/dkrizic/feature/cli/repository/feature/v1"
	"github.com/urfave/cli/v3"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func FeatureClient(cmd *cli.Command) (feature.FeatureClient, error) {
	endpoint := cmd.String(constant.Endpoint)

	gc, err := grpc.NewClient(endpoint,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
	)
	if err != nil {
		return nil, err
	}

	fc := feature.NewFeatureClient(gc)
	return fc, nil
}
