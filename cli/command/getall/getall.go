package getall

import (
	"context"

	"github.com/dkrizic/feature/cli/constant"
	feature "github.com/dkrizic/feature/cli/repository/feature/v1"
	"github.com/urfave/cli/v3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"
)

func GetAll(ctx context.Context, cmd *cli.Command) error {
	endpoint := cmd.String(constant.Endpoint)

	gc, err := grpc.NewClient(endpoint, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}
	defer gc.Close()

	fc := feature.NewFeatureClient(gc)
	all, err := fc.GetAll(ctx, &emptypb.Empty{})
	for {
		kv, err := all.Recv()
		if err != nil {
			break
		}
		cmd.Writer.Write([]byte(kv.Key + ": " + kv.Value + "\n"))
	}
	return nil
}
