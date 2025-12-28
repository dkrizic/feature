package feature

import (
	"context"
	"log/slog"

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
	count := 0
	for _, kv := range values {
		err := stream.Send(&featurev1.KeyValue{
			Key:   kv.Key,
			Value: kv.Value,
		})
		if err != nil {
			return err
		}
		count++
	}
	slog.Info("GetAll completed", "count", count)
	return nil
}

func (fs *FeatureService) PreSet(ctx context.Context, kv *featurev1.KeyValue) (*emptypb.Empty, error) {
	err := fs.persistence.PreSet(ctx, persistence.KeyValue{
		Key:   kv.Key,
		Value: kv.Value,
	})
	if err != nil {
		return nil, err
	}
	slog.Info("PreSet completed", "key", kv.Key, "value", kv.Value)
	return &emptypb.Empty{}, nil
}

func (fs *FeatureService) Set(ctx context.Context, kv *featurev1.KeyValue) (*emptypb.Empty, error) {
	err := fs.persistence.Set(ctx, persistence.KeyValue{
		Key:   kv.Key,
		Value: kv.Value,
	})
	if err != nil {
		return nil, err
	}
	slog.Info("Set completed", "key", kv.Key, "value", kv.Value)
	return &emptypb.Empty{}, nil
}

func (fs *FeatureService) Get(ctx context.Context, kv *featurev1.Key) (*featurev1.Value, error) {
	result, err := fs.persistence.Get(ctx, kv.Name)
	if err != nil {
		return nil, err
	}
	slog.Info("Get completed", "key", kv.Name, "value", result)
	return &featurev1.Value{
		Name: result.Value,
	}, nil
}

func (fs *FeatureService) Delete(ctx context.Context, kv *featurev1.Key) (*emptypb.Empty, error) {
	err := fs.persistence.Delete(ctx, kv.Name)
	if err != nil {
		return nil, err
	}
	slog.Info("Delete completed", "key", kv.Name)
	return &emptypb.Empty{}, nil
}
