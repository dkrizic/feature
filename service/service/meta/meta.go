package meta

// implement this interface from api/meta/gen package

import (
	"context"
	"github.com/dkrizic/feature/service/meta"
	metav1 "github.com/dkrizic/feature/service/service/meta/v1"
)

type MetaService struct {
	metav1.UnimplementedMetaServer
}

func New() *MetaService {
	return &MetaService{}
}

func (ms MetaService) Meta(ctx context.Context, req *metav1.MetaRequest) (*metav1.MetaResponse, error) {
	//	ctx, span := telemetry.Tracer().Start(ctx, "meta.MetaService.Meta")
	//	defer span.End()

	//	span.SetAttributes(
	//		attribute.String("service.name", meta.ServiceName),
	//		attribute.String("service.version", meta.Version),
	//	)

	return &metav1.MetaResponse{
		ServiceName: meta.Service,
		Version:     meta.Version,
	}, nil
}
