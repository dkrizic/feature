package application

import (
	"context"
	"os"
	"testing"

	"github.com/dkrizic/feature/service/constant"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v3"
)

func TestManagerLegacyMode(t *testing.T) {
	// Test legacy single-application mode
	ctx := context.Background()
	
	// Create a command with legacy configuration
	cmd := &cli.Command{}
	cmd.Flags = []cli.Flag{
		&cli.StringFlag{Name: constant.StorageType, Value: constant.StorageTypeInMemory},
		&cli.StringFlag{Name: constant.ConfigMapName, Value: ""},
		&cli.StringFlag{Name: constant.Editable, Value: ""},
		&cli.StringSliceFlag{Name: constant.PreSet, Value: []string{"KEY1=value1"}},
		&cli.BoolFlag{Name: constant.RestartEnabled, Value: false},
		&cli.StringFlag{Name: constant.RestartType, Value: "deployment"},
		&cli.StringFlag{Name: constant.RestartName, Value: ""},
		&cli.BoolFlag{Name: constant.NotificationEnabled, Value: false},
		&cli.StringFlag{Name: constant.NotificationType, Value: constant.NotificationTypeLog},
	}
	
	manager := NewManager()
	err := manager.LoadFromConfig(ctx, cmd)
	assert.NoError(t, err)
	
	// Should have one default application
	app, err := manager.GetApplication("")
	assert.NoError(t, err)
	assert.Equal(t, "default", app.Name)
	assert.Equal(t, constant.StorageTypeInMemory, app.StorageType)
	
	// Default application should be "default"
	assert.Equal(t, "default", manager.GetDefaultApplication())
	
	// Should have one application in the list
	apps := manager.ListApplications()
	assert.Len(t, apps, 1)
}

func TestManagerMultiApplicationMode(t *testing.T) {
	// Test multi-application mode
	ctx := context.Background()
	
	// Set up environment variables for multi-application mode
	os.Setenv("APPLICATIONS", "app1,app2")
	os.Setenv("DEFAULT_APPLICATION", "app1")
	
	// app1 configuration
	os.Setenv("APP1_NAMESPACE", "namespace1")
	os.Setenv("APP1_STORAGE_TYPE", "inmemory")
	os.Setenv("APP1_PRESET", "KEY1=value1")
	os.Setenv("APP1_EDITABLE", "KEY1")
	
	// app2 configuration
	os.Setenv("APP2_NAMESPACE", "namespace2")
	os.Setenv("APP2_STORAGE_TYPE", "inmemory")
	os.Setenv("APP2_PRESET", "KEY2=value2")
	os.Setenv("APP2_EDITABLE", "KEY2")
	
	defer func() {
		// Clean up environment variables
		os.Unsetenv("APPLICATIONS")
		os.Unsetenv("DEFAULT_APPLICATION")
		os.Unsetenv("APP1_NAMESPACE")
		os.Unsetenv("APP1_STORAGE_TYPE")
		os.Unsetenv("APP1_PRESET")
		os.Unsetenv("APP1_EDITABLE")
		os.Unsetenv("APP2_NAMESPACE")
		os.Unsetenv("APP2_STORAGE_TYPE")
		os.Unsetenv("APP2_PRESET")
		os.Unsetenv("APP2_EDITABLE")
	}()
	
	// Create a command with notification settings
	cmd := &cli.Command{}
	cmd.Flags = []cli.Flag{
		&cli.BoolFlag{Name: constant.NotificationEnabled, Value: false},
		&cli.StringFlag{Name: constant.NotificationType, Value: constant.NotificationTypeLog},
	}
	
	manager := NewManager()
	err := manager.LoadFromConfig(ctx, cmd)
	assert.NoError(t, err)
	
	// Should have two applications
	apps := manager.ListApplications()
	assert.Len(t, apps, 2)
	
	// Check app1
	app1, err := manager.GetApplication("app1")
	assert.NoError(t, err)
	assert.Equal(t, "app1", app1.Name)
	assert.Equal(t, "namespace1", app1.Namespace)
	assert.Equal(t, constant.StorageTypeInMemory, app1.StorageType)
	assert.Len(t, app1.EditableList, 1)
	assert.Equal(t, "KEY1", app1.EditableList[0])
	
	// Check app2
	app2, err := manager.GetApplication("app2")
	assert.NoError(t, err)
	assert.Equal(t, "app2", app2.Name)
	assert.Equal(t, "namespace2", app2.Namespace)
	assert.Equal(t, constant.StorageTypeInMemory, app2.StorageType)
	assert.Len(t, app2.EditableList, 1)
	assert.Equal(t, "KEY2", app2.EditableList[0])
	
	// Default application should be app1
	assert.Equal(t, "app1", manager.GetDefaultApplication())
	
	// Getting application with empty name should return default
	app, err := manager.GetApplication("")
	assert.NoError(t, err)
	assert.Equal(t, "app1", app.Name)
}

func TestManagerApplicationNotFound(t *testing.T) {
	// Test that getting a non-existent application returns an error
	ctx := context.Background()
	
	cmd := &cli.Command{}
	cmd.Flags = []cli.Flag{
		&cli.StringFlag{Name: constant.StorageType, Value: constant.StorageTypeInMemory},
		&cli.StringFlag{Name: constant.ConfigMapName, Value: ""},
		&cli.StringFlag{Name: constant.Editable, Value: ""},
		&cli.StringSliceFlag{Name: constant.PreSet, Value: []string{}},
		&cli.BoolFlag{Name: constant.RestartEnabled, Value: false},
		&cli.StringFlag{Name: constant.RestartType, Value: "deployment"},
		&cli.StringFlag{Name: constant.RestartName, Value: ""},
		&cli.BoolFlag{Name: constant.NotificationEnabled, Value: false},
		&cli.StringFlag{Name: constant.NotificationType, Value: constant.NotificationTypeLog},
	}
	
	manager := NewManager()
	err := manager.LoadFromConfig(ctx, cmd)
	assert.NoError(t, err)
	
	// Try to get a non-existent application
	_, err = manager.GetApplication("non-existent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "application not found")
}
