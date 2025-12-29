package getall

import (
	"context"

	"github.com/dkrizic/feature/cli/command"
	"github.com/urfave/cli/v3"
	"google.golang.org/protobuf/types/known/emptypb"
)

func GetAll(ctx context.Context, cmd *cli.Command) error {
	fc, err := command.FeatureClient(cmd)
	if err != nil {
		return err
	}

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
