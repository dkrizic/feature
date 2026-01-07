package feature

import (
	"context"
	"log/slog"

	"github.com/dkrizic/feature/service/service/feature/v1"
	"github.com/dkrizic/feature/service/service/persistence"
	"github.com/dkrizic/feature/service/telemetry/localmetrics"
	"go.opentelemetry.io/otel"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type FeatureService struct {
	featurev1.UnimplementedFeatureServer
	persistence persistence.Persistence
}

func NewFeatureService(p persistence.Persistence) (*FeatureService, error) {
	err := localmetrics.New()
	if err != nil {
		slog.Error("Failed to initialize local metrics", "error", err)
	}
	return &FeatureService{
		persistence: p,
	}, nil
}

func (fs *FeatureService) GetAll(empty *emptypb.Empty, stream grpc.ServerStreamingServer[featurev1.KeyValue]) error {
	ctx, span := otel.Tracer("feature/service").Start(stream.Context(), "GetAll")
	defer span.End()

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
	count := len(values)
	localmetrics.ActiveGauge().Record(ctx, int64(count))
	localmetrics.GetAllCounter().Add(ctx, 1)

	slog.InfoContext(ctx, "GetAll completed", "count", count)
	return nil
}

func (fs *FeatureService) PreSet(ctx context.Context, kv *featurev1.KeyValue) (*emptypb.Empty, error) {
	ctx, span := otel.Tracer("feature/service").Start(ctx, "PreSet")
	defer span.End()

	err := fs.persistence.PreSet(ctx, persistence.KeyValue{
		Key:   kv.Key,
		Value: kv.Value,
	})
	if err != nil {
		return nil, err
	}
	slog.InfoContext(ctx, "PreSet completed", "key", kv.Key, "value", kv.Value)
	count, err := fs.persistence.Count(ctx)
	if err != nil {
		return nil, err
	}
	localmetrics.ActiveGauge().Record(ctx, int64(count))
	localmetrics.PresetCounter().Add(ctx, 1)
	return &emptypb.Empty{}, nil
}

func (fs *FeatureService) Set(ctx context.Context, kv *featurev1.KeyValue) (*emptypb.Empty, error) {
	ctx, span := otel.Tracer("feature/service").Start(ctx, "Set")
	defer span.End()

	err := fs.persistence.Set(ctx, persistence.KeyValue{
		Key:   kv.Key,
		Value: kv.Value,
	})
	if err != nil {
		return nil, err
	}
	slog.InfoContext(ctx, "Set completed", "key", kv.Key, "value", kv.Value)
	count, err := fs.persistence.Count(ctx)
	if err != nil {
		return nil, err
	}
	localmetrics.ActiveGauge().Record(ctx, int64(count))
	localmetrics.SetCounter().Add(ctx, 1)
	return &emptypb.Empty{}, nil
}

func (fs *FeatureService) Get(ctx context.Context, kv *featurev1.Key) (*featurev1.Value, error) {
	ctx, span := otel.Tracer("feature/service").Start(ctx, "Get")
	defer span.End()

	result, err := fs.persistence.Get(ctx, kv.Name)
	if err != nil {
		return nil, err
	}
	slog.InfoContext(ctx, "Get completed", "key", kv.Name, "value", result)
	count, err := fs.persistence.Count(ctx)
	if err != nil {
		return nil, err
	}
	localmetrics.ActiveGauge().Record(ctx, int64(count))
	localmetrics.GetCounter().Add(ctx, 1)
	return &featurev1.Value{
		Name: result.Value,
	}, nil
}

func (fs *FeatureService) Delete(ctx context.Context, kv *featurev1.Key) (*emptypb.Empty, error) {
	ctx, span := otel.Tracer("feature/service").Start(ctx, "Delete")
	defer span.End()

	err := fs.persistence.Delete(ctx, kv.Name)
	if err != nil {
		return nil, err
	}
	slog.InfoContext(ctx, "Delete completed", "key", kv.Name)
	count, err := fs.persistence.Count(ctx)
	if err != nil {
		return nil, err
	}
	localmetrics.ActiveGauge().Record(ctx, int64(count))
	localmetrics.DeleteCounter().Add(ctx, 1)
	return &emptypb.Empty{}, nil
}
