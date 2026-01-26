package feature

import (
	"context"
	"log/slog"
	"strings"

	"github.com/dkrizic/feature/service/service/application"
	"github.com/dkrizic/feature/service/service/feature/v1"
	"github.com/dkrizic/feature/service/service/persistence"
	"github.com/dkrizic/feature/service/telemetry/localmetrics"
	"go.opentelemetry.io/otel"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type FeatureService struct {
	featurev1.UnimplementedFeatureServer
	appManager     *application.Manager
	// Legacy fields for backward compatibility
	persistence    persistence.Persistence
	editableFields map[string]bool // map of editable field names, empty means all are editable
}

// parseEditableFields parses a comma-separated list of field names and returns a map
func parseEditableFields(editableFieldsStr string) map[string]bool {
	editableFields := make(map[string]bool)
	if editableFieldsStr != "" {
		fields := strings.Split(editableFieldsStr, ",")
		for _, field := range fields {
			field = strings.TrimSpace(field)
			if field != "" {
				editableFields[field] = true
			}
		}
	}
	return editableFields
}

func NewFeatureService(p persistence.Persistence, editableFieldsStr string) (*FeatureService, error) {
	err := localmetrics.New()
	if err != nil {
		slog.Error("Failed to initialize local metrics", "error", err)
	}

	// Parse editable fields
	editableFields := parseEditableFields(editableFieldsStr)
	
	if len(editableFields) > 0 {
		slog.Info("Editable fields configured", "fields", editableFieldsStr, "count", len(editableFields))
	} else {
		slog.Info("All fields are editable (no restrictions)")
	}

	return &FeatureService{
		persistence:    p,
		editableFields: editableFields,
	}, nil
}

// NewFeatureServiceWithAppManager creates a new feature service with multi-application support
func NewFeatureServiceWithAppManager(appManager *application.Manager) (*FeatureService, error) {
	err := localmetrics.New()
	if err != nil {
		slog.Error("Failed to initialize local metrics", "error", err)
	}

	return &FeatureService{
		appManager: appManager,
	}, nil
}

// isEditable checks if a field is editable
func (fs *FeatureService) isEditable(key string) bool {
	// If editableFields is empty, all fields are editable
	if len(fs.editableFields) == 0 {
		return true
	}
	// Otherwise, check if the key is in the editable list
	return fs.editableFields[key]
}

// isLegacyMode returns true if the service is running in legacy single-application mode
func (fs *FeatureService) isLegacyMode() bool {
	return fs.appManager == nil
}

// getAppPersistence returns the persistence and editable fields for a specific application
func (fs *FeatureService) getAppPersistence(ctx context.Context, appName string) (persistence.Persistence, map[string]bool, error) {
	// If using legacy mode (single application)
	if fs.isLegacyMode() {
		return fs.persistence, fs.editableFields, nil
	}
	
	// Multi-application mode
	app, err := fs.appManager.GetApplication(appName)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to get application", "application", appName, "error", err)
		return nil, nil, status.Errorf(codes.NotFound, "application not found: %s", appName)
	}
	
	editableFields := make(map[string]bool)
	for _, field := range app.EditableList {
		editableFields[field] = true
	}
	
	return app.Persistence, editableFields, nil
}

func (fs *FeatureService) GetAll(empty *emptypb.Empty, stream grpc.ServerStreamingServer[featurev1.KeyValue]) error {
	ctx, span := otel.Tracer("feature/service").Start(stream.Context(), "GetAll")
	defer span.End()

	// For backward compatibility, if no application is specified and we're in legacy mode,
	// use the default persistence. Otherwise, we would need application context from metadata
	// For now, we'll get all from the default application
	var appName string
	if !fs.isLegacyMode() {
		appName = fs.appManager.GetDefaultApplication()
	}
	
	pers, editableFields, err := fs.getAppPersistence(ctx, appName)
	if err != nil {
		return err
	}

	values, err := pers.GetAll(ctx)
	if err != nil {
		return err
	}
	for _, kv := range values {
		editable := true
		if len(editableFields) > 0 {
			editable = editableFields[kv.Key]
		}
		err := stream.Send(&featurev1.KeyValue{
			Key:         kv.Key,
			Value:       kv.Value,
			Editable:    editable,
			Application: appName,
		})
		if err != nil {
			return err
		}
	}
	count := len(values)
	localmetrics.ActiveGauge().Record(ctx, int64(count))
	localmetrics.GetAllCounter().Add(ctx, 1)

	slog.InfoContext(ctx, "GetAll completed", "count", count, "application", appName)
	return nil
}

func (fs *FeatureService) PreSet(ctx context.Context, kv *featurev1.KeyValue) (*emptypb.Empty, error) {
	ctx, span := otel.Tracer("feature/service").Start(ctx, "PreSet")
	defer span.End()

	pers, _, err := fs.getAppPersistence(ctx, kv.Application)
	if err != nil {
		return nil, err
	}

	err = pers.PreSet(ctx, persistence.KeyValue{
		Key:   kv.Key,
		Value: kv.Value,
	})
	if err != nil {
		return nil, err
	}
	slog.InfoContext(ctx, "PreSet completed", "key", kv.Key, "value", kv.Value, "application", kv.Application)
	count, err := pers.Count(ctx)
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

	pers, editableFields, err := fs.getAppPersistence(ctx, kv.Application)
	if err != nil {
		return nil, err
	}

	// If editable fields are configured (not empty), additional restrictions apply
	if len(editableFields) > 0 {
		// Check if the field already exists by getting all fields
		allFields, err := pers.GetAll(ctx)
		if err != nil {
			return nil, err
		}
		
		fieldExists := false
		for _, field := range allFields {
			if field.Key == kv.Key {
				fieldExists = true
				break
			}
		}
		
		// If field doesn't exist, creating new fields is not allowed
		if !fieldExists {
			slog.WarnContext(ctx, "Attempt to create new field when editable restrictions are active", "key", kv.Key, "application", kv.Application)
			return nil, status.Errorf(codes.PermissionDenied, "creating new fields is not allowed when editable restrictions are active")
		}
		
		// Check if the existing field is editable
		if !editableFields[kv.Key] {
			slog.WarnContext(ctx, "Attempt to set non-editable field", "key", kv.Key, "application", kv.Application)
			return nil, status.Errorf(codes.PermissionDenied, "field '%s' is not editable", kv.Key)
		}
	}

	err = pers.Set(ctx, persistence.KeyValue{
		Key:   kv.Key,
		Value: kv.Value,
	})
	if err != nil {
		return nil, err
	}
	slog.InfoContext(ctx, "Set completed", "key", kv.Key, "value", kv.Value, "application", kv.Application)
	count, err := pers.Count(ctx)
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

	pers, _, err := fs.getAppPersistence(ctx, kv.Application)
	if err != nil {
		return nil, err
	}

	result, err := pers.Get(ctx, kv.Name)
	if err != nil {
		return nil, err
	}
	slog.InfoContext(ctx, "Get completed", "key", kv.Name, "value", result, "application", kv.Application)
	count, err := pers.Count(ctx)
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

	pers, editableFields, err := fs.getAppPersistence(ctx, kv.Application)
	if err != nil {
		return nil, err
	}

	// If editable fields are configured (not empty), deletion is not allowed
	if len(editableFields) > 0 {
		slog.WarnContext(ctx, "Attempt to delete field when editable restrictions are active", "key", kv.Name, "application", kv.Application)
		return nil, status.Errorf(codes.PermissionDenied, "deleting fields is not allowed when editable restrictions are active")
	}

	err = pers.Delete(ctx, kv.Name)
	if err != nil {
		return nil, err
	}
	slog.InfoContext(ctx, "Delete completed", "key", kv.Name, "application", kv.Application)
	count, err := pers.Count(ctx)
	if err != nil {
		return nil, err
	}
	localmetrics.ActiveGauge().Record(ctx, int64(count))
	localmetrics.DeleteCounter().Add(ctx, 1)
	return &emptypb.Empty{}, nil
}

// GetApplications returns a list of all configured applications
func (fs *FeatureService) GetApplications(req *featurev1.ApplicationsRequest, stream grpc.ServerStreamingServer[featurev1.Application]) error {
	ctx, span := otel.Tracer("feature/service").Start(stream.Context(), "GetApplications")
	defer span.End()

	if fs.isLegacyMode() {
		// Legacy mode - return default application
		err := stream.Send(&featurev1.Application{
			Name:        "default",
			Namespace:   "default",
			StorageType: "inmemory",
		})
		if err != nil {
			return err
		}
		slog.InfoContext(ctx, "GetApplications completed (legacy mode)", "count", 1)
		return nil
	}

	// Multi-application mode
	apps := fs.appManager.ListApplications()
	for _, app := range apps {
		err := stream.Send(&featurev1.Application{
			Name:        app.Name,
			Namespace:   app.Namespace,
			StorageType: app.StorageType,
		})
		if err != nil {
			return err
		}
	}

	slog.InfoContext(ctx, "GetApplications completed", "count", len(apps))
	return nil
}
