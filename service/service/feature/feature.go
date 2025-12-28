package feature

import (
	"context"

	"github.com/dkrizic/feature/service/service/feature/featurev1"
	"github.com/dkrizic/feature/service/service/persistence"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type FeatureService struct {
	featurev1.UnimplementedFeatureServer
	persistence persistence.Persistence
}

func NewFeatureService(p persistence.Persistence) *FeatureService {
	return &FeatureService{
		persistence: p,
	}
}

func (fs *FeatureService) GetAll(empty *emptypb.Empty, stream grpc.ServerStreamingServer[featurev1.KeyValue]) error {
	ctx := stream.Context()
	values, err := fs.persistence.GetAll(ctx)
	if err != nil {
		return err
	}
	for _, kv := range values {
		err := stream.Send(&featurev1.KeyValue{
			Key:   kv.Key,
			Value: kv.Value,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func (fs *FeatureService) PreSet(context.Context, *featurev1.KeyValue) (*emptypb.Empty, error) {
	return nil, nil
}

func (fs *FeatureService) Set(context.Context, *featurev1.KeyValue) (*emptypb.Empty, error) {
	return nil, nil
}

func (fs *FeatureService) Get(context.Context, *featurev1.Key) (*featurev1.Value, error) {
	return nil, nil
}
