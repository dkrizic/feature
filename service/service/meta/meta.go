package meta

// implement this interface from api/meta/gen package

import (
	"context"

	"github.com/dkrizic/feature/service/meta"
	metav1 "github.com/dkrizic/feature/service/service/meta/v1"
	"go.opentelemetry.io/otel"
)

type MetaService struct {
	metav1.UnimplementedMetaServer
	authenticationRequired bool
}

func New(authenticationRequired bool) *MetaService {
	return &MetaService{
		authenticationRequired: authenticationRequired,
	}
}

func (ms MetaService) Meta(ctx context.Context, req *metav1.MetaRequest) (*metav1.MetaResponse, error) {
	ctx, span := otel.Tracer("feature/service").Start(ctx, "Handle")
	defer span.End()

	return &metav1.MetaResponse{
		ServiceName:            meta.Service,
		Version:                meta.Version,
		AuthenticationRequired: ms.authenticationRequired,
	}, nil
}
