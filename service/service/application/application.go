package application

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/dkrizic/feature/service/constant"
	nf "github.com/dkrizic/feature/service/notifier/factory"
	"github.com/dkrizic/feature/service/service/persistence"
	"github.com/dkrizic/feature/service/service/persistence/configmap"
	"github.com/dkrizic/feature/service/service/persistence/inmemory"
	"github.com/dkrizic/feature/service/service/persistence/notifying"
	"github.com/urfave/cli/v3"
)

// Application represents a single application configuration
type Application struct {
	Name         string
	Namespace    string
	StorageType  string
	ConfigMap    ConfigMapConfig
	Workload     WorkloadConfig
	Persistence  persistence.Persistence
	EditableList []string
}

// ConfigMapConfig holds ConfigMap-specific configuration
type ConfigMapConfig struct {
	Name     string
	Preset   []string
	Editable string
}

// WorkloadConfig holds workload restart configuration
type WorkloadConfig struct {
	Enabled bool
	Type    string
	Name    string
}

// Manager manages multiple applications
type Manager struct {
	applications      map[string]*Application
	defaultApplication string
}

// NewManager creates a new application manager
func NewManager() *Manager {
	return &Manager{
		applications: make(map[string]*Application),
	}
}

// LoadFromConfig loads applications from CLI command configuration
func (m *Manager) LoadFromConfig(ctx context.Context, cmd *cli.Command) error {
	// Check if we're using the old single-application config or new multi-application config
	applicationsStr := os.Getenv("APPLICATIONS")
	
	if applicationsStr == "" {
		// Legacy single-application mode
		return m.loadLegacyConfig(ctx, cmd)
	}
	
	// Multi-application mode
	return m.loadMultiApplicationConfig(ctx, cmd, applicationsStr)
}

// loadLegacyConfig loads a single application from the old configuration format
func (m *Manager) loadLegacyConfig(ctx context.Context, cmd *cli.Command) error {
	slog.InfoContext(ctx, "Loading legacy single-application configuration")
	
	storageType := cmd.String(constant.StorageType)
	configMapName := cmd.String(constant.ConfigMapName)
	editable := cmd.String(constant.Editable)
	preset := cmd.StringSlice(constant.PreSet)
	restartEnabled := cmd.Bool(constant.RestartEnabled)
	restartType := cmd.String(constant.RestartType)
	restartName := cmd.String(constant.RestartName)
	
	namespace := os.Getenv("POD_NAMESPACE")
	if namespace == "" {
		namespace = "default"
	}
	
	app := &Application{
		Name:      "default",
		Namespace: namespace,
		StorageType: storageType,
		ConfigMap: ConfigMapConfig{
			Name:     configMapName,
			Preset:   preset,
			Editable: editable,
		},
		Workload: WorkloadConfig{
			Enabled: restartEnabled,
			Type:    restartType,
			Name:    restartName,
		},
	}
	
	// Parse editable fields
	if editable != "" {
		app.EditableList = strings.Split(editable, ",")
		for i := range app.EditableList {
			app.EditableList[i] = strings.TrimSpace(app.EditableList[i])
		}
	}
	
	// Create persistence
	notifier, err := nf.NewNotifier(ctx, cmd)
	if err != nil {
		return err
	}
	
	switch storageType {
	case constant.StorageTypeInMemory:
		slog.InfoContext(ctx, "In-memory storage selected for application", "application", app.Name)
		app.Persistence = notifying.NewNotifyingPersistence(
			inmemory.NewInMemoryPersistence(), notifier,
		)
	case constant.StorageTypeConfigMap:
		slog.InfoContext(ctx, "ConfigMap storage selected for application", "application", app.Name, "configmap", configMapName)
		app.Persistence = notifying.NewNotifyingPersistence(
			configmap.NewConfigMapPersistence(configMapName), notifier,
		)
	default:
		return fmt.Errorf("invalid storage type: %s", storageType)
	}
	
	m.applications[app.Name] = app
	m.defaultApplication = app.Name
	
	slog.InfoContext(ctx, "Loaded application", "name", app.Name, "namespace", app.Namespace, "storage", app.StorageType)
	
	return nil
}

// loadMultiApplicationConfig loads multiple applications from environment variables
func (m *Manager) loadMultiApplicationConfig(ctx context.Context, cmd *cli.Command, applicationsStr string) error {
	slog.InfoContext(ctx, "Loading multi-application configuration")
	
	appNames := strings.Split(applicationsStr, ",")
	
	// Get default application
	defaultApp := os.Getenv("DEFAULT_APPLICATION")
	if defaultApp == "" && len(appNames) > 0 {
		defaultApp = strings.TrimSpace(appNames[0])
	}
	m.defaultApplication = defaultApp
	
	notifier, err := nf.NewNotifier(ctx, cmd)
	if err != nil {
		return err
	}
	
	for _, appName := range appNames {
		appName = strings.TrimSpace(appName)
		if appName == "" {
			continue
		}
		
		// Load application configuration from environment variables
		prefix := strings.ToUpper(strings.ReplaceAll(appName, "-", "_"))
		
		namespace := os.Getenv(prefix + "_NAMESPACE")
		if namespace == "" {
			namespace = "default"
		}
		
		storageType := os.Getenv(prefix + "_STORAGE_TYPE")
		if storageType == "" {
			storageType = constant.StorageTypeInMemory
		}
		
		configMapName := os.Getenv(prefix + "_CONFIGMAP_NAME")
		presetStr := os.Getenv(prefix + "_PRESET")
		editable := os.Getenv(prefix + "_EDITABLE")
		
		var preset []string
		if presetStr != "" {
			preset = strings.Split(presetStr, ",")
		}
		
		restartEnabledStr := os.Getenv(prefix + "_RESTART_ENABLED")
		restartEnabled := restartEnabledStr == "true"
		restartType := os.Getenv(prefix + "_RESTART_TYPE")
		if restartType == "" {
			restartType = "deployment"
		}
		restartName := os.Getenv(prefix + "_RESTART_NAME")
		
		app := &Application{
			Name:        appName,
			Namespace:   namespace,
			StorageType: storageType,
			ConfigMap: ConfigMapConfig{
				Name:     configMapName,
				Preset:   preset,
				Editable: editable,
			},
			Workload: WorkloadConfig{
				Enabled: restartEnabled,
				Type:    restartType,
				Name:    restartName,
			},
		}
		
		// Parse editable fields
		if editable != "" {
			app.EditableList = strings.Split(editable, ",")
			for i := range app.EditableList {
				app.EditableList[i] = strings.TrimSpace(app.EditableList[i])
			}
		}
		
		// Create persistence for this application
		switch storageType {
		case constant.StorageTypeInMemory:
			slog.InfoContext(ctx, "In-memory storage selected for application", "application", app.Name)
			app.Persistence = notifying.NewNotifyingPersistence(
				inmemory.NewInMemoryPersistence(), notifier,
			)
		case constant.StorageTypeConfigMap:
			if configMapName == "" {
				return fmt.Errorf("configmap name is required for application %s when using configmap storage", appName)
			}
			slog.InfoContext(ctx, "ConfigMap storage selected for application", "application", app.Name, "configmap", configMapName, "namespace", namespace)
			app.Persistence = notifying.NewNotifyingPersistence(
				configmap.NewConfigMapPersistence(configMapName), notifier,
			)
		default:
			return fmt.Errorf("invalid storage type for application %s: %s", appName, storageType)
		}
		
		m.applications[app.Name] = app
		
		slog.InfoContext(ctx, "Loaded application", "name", app.Name, "namespace", app.Namespace, "storage", app.StorageType)
	}
	
	if len(m.applications) == 0 {
		return fmt.Errorf("no applications configured")
	}
	
	slog.InfoContext(ctx, "Application configuration complete", "count", len(m.applications), "default", m.defaultApplication)
	
	return nil
}

// GetApplication returns an application by name
func (m *Manager) GetApplication(name string) (*Application, error) {
	if name == "" {
		name = m.defaultApplication
	}
	
	app, ok := m.applications[name]
	if !ok {
		return nil, fmt.Errorf("application not found: %s", name)
	}
	
	return app, nil
}

// GetDefaultApplication returns the default application name
func (m *Manager) GetDefaultApplication() string {
	return m.defaultApplication
}

// ListApplications returns all configured applications
func (m *Manager) ListApplications() []*Application {
	apps := make([]*Application, 0, len(m.applications))
	for _, app := range m.applications {
		apps = append(apps, app)
	}
	return apps
}

// PreSetApplication applies preset values for a specific application
func (m *Manager) PreSetApplication(ctx context.Context, appName string) error {
	app, err := m.GetApplication(appName)
	if err != nil {
		return err
	}
	
	for _, kv := range app.ConfigMap.Preset {
		parts := strings.SplitN(kv, "=", 2)
		if len(parts) != 2 {
			slog.WarnContext(ctx, "Invalid preset format, expected key=value", "preset", kv, "application", appName)
			continue
		}
		key := parts[0]
		value := parts[1]
		slog.InfoContext(ctx, "Pre-setting key-value", "key", key, "value", value, "application", appName)
		err := app.Persistence.PreSet(ctx, persistence.KeyValue{
			Key:   key,
			Value: value,
		})
		if err != nil {
			slog.ErrorContext(ctx, "Failed to pre-set key-value", "key", key, "value", value, "application", appName, "error", err)
			return fmt.Errorf("failed to pre-set key-value for application %s: %w", appName, err)
		}
	}
	
	return nil
}
