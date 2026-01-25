package workload

import (
"context"
"fmt"
"log/slog"
"time"

workloadv1 "github.com/dkrizic/feature/service/service/workload/v1"
metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
"k8s.io/client-go/kubernetes"
"k8s.io/client-go/rest"
)

// WorkloadService implements the Workload gRPC service
type WorkloadService struct {
	workloadv1.UnimplementedWorkloadServer
	clientset      *kubernetes.Clientset
	namespace      string
	restartEnabled bool
	restartType    workloadv1.WorkloadType
	restartName    string
}

// NewWorkloadService creates a new workload service
func NewWorkloadService(namespace string, restartEnabled bool, restartType workloadv1.WorkloadType, restartName string) (*WorkloadService, error) {
	// Create in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to create in-cluster config: %w", err)
	}

	// Create Kubernetes clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes clientset: %w", err)
	}

	return &WorkloadService{
		clientset:      clientset,
		namespace:      namespace,
		restartEnabled: restartEnabled,
		restartType:    restartType,
		restartName:    restartName,
	}, nil
}

// RestartWorkload performs a rollout restart on the specified workload
func (s *WorkloadService) RestartWorkload(ctx context.Context, req *workloadv1.RestartRequest) (*workloadv1.RestartResponse, error) {
slog.InfoContext(ctx, "Received restart request", "type", req.Type.String(), "name", req.Name, "namespace", req.Namespace)

// Use request namespace if provided, otherwise use service namespace
namespace := req.Namespace
if namespace == "" {
namespace = s.namespace
}

// Validate input
if req.Name == "" {
return &workloadv1.RestartResponse{
Success: false,
Message: "Workload name is required",
}, nil
}

if req.Type == workloadv1.WorkloadType_WORKLOAD_TYPE_UNSPECIFIED {
return &workloadv1.RestartResponse{
Success: false,
Message: "Workload type must be specified",
}, nil
}

// Perform restart based on workload type
var err error
switch req.Type {
case workloadv1.WorkloadType_WORKLOAD_TYPE_DEPLOYMENT:
err = s.restartDeployment(ctx, namespace, req.Name)
case workloadv1.WorkloadType_WORKLOAD_TYPE_STATEFULSET:
err = s.restartStatefulSet(ctx, namespace, req.Name)
case workloadv1.WorkloadType_WORKLOAD_TYPE_DAEMONSET:
err = s.restartDaemonSet(ctx, namespace, req.Name)
default:
return &workloadv1.RestartResponse{
Success: false,
Message: fmt.Sprintf("Unsupported workload type: %s", req.Type.String()),
}, nil
}

if err != nil {
slog.ErrorContext(ctx, "Failed to restart workload", "type", req.Type.String(), "name", req.Name, "namespace", namespace, "error", err)
return &workloadv1.RestartResponse{
Success: false,
Message: fmt.Sprintf("Failed to restart %s: %v", req.Type.String(), err),
}, nil
}

slog.InfoContext(ctx, "Successfully restarted workload", "type", req.Type.String(), "name", req.Name, "namespace", namespace)
return &workloadv1.RestartResponse{
Success: true,
Message: fmt.Sprintf("Successfully restarted %s/%s", req.Type.String(), req.Name),
}, nil
}

// restartDeployment performs a rollout restart on a deployment
func (s *WorkloadService) restartDeployment(ctx context.Context, namespace, name string) error {
deploymentsClient := s.clientset.AppsV1().Deployments(namespace)

// Get the deployment
deployment, err := deploymentsClient.Get(ctx, name, metav1.GetOptions{})
if err != nil {
return fmt.Errorf("failed to get deployment: %w", err)
}

// Set the restart annotation
if deployment.Spec.Template.Annotations == nil {
deployment.Spec.Template.Annotations = make(map[string]string)
}
deployment.Spec.Template.Annotations["kubectl.kubernetes.io/restartedAt"] = time.Now().Format(time.RFC3339)

// Update the deployment
_, err = deploymentsClient.Update(ctx, deployment, metav1.UpdateOptions{})
if err != nil {
return fmt.Errorf("failed to update deployment: %w", err)
}

return nil
}

// restartStatefulSet performs a rollout restart on a statefulset
func (s *WorkloadService) restartStatefulSet(ctx context.Context, namespace, name string) error {
statefulSetsClient := s.clientset.AppsV1().StatefulSets(namespace)

// Get the statefulset
statefulSet, err := statefulSetsClient.Get(ctx, name, metav1.GetOptions{})
if err != nil {
return fmt.Errorf("failed to get statefulset: %w", err)
}

// Set the restart annotation
if statefulSet.Spec.Template.Annotations == nil {
statefulSet.Spec.Template.Annotations = make(map[string]string)
}
statefulSet.Spec.Template.Annotations["kubectl.kubernetes.io/restartedAt"] = time.Now().Format(time.RFC3339)

// Update the statefulset
_, err = statefulSetsClient.Update(ctx, statefulSet, metav1.UpdateOptions{})
if err != nil {
return fmt.Errorf("failed to update statefulset: %w", err)
}

return nil
}

// restartDaemonSet performs a rollout restart on a daemonset
func (s *WorkloadService) restartDaemonSet(ctx context.Context, namespace, name string) error {
daemonSetsClient := s.clientset.AppsV1().DaemonSets(namespace)

// Get the daemonset
daemonSet, err := daemonSetsClient.Get(ctx, name, metav1.GetOptions{})
if err != nil {
return fmt.Errorf("failed to get daemonset: %w", err)
}

// Set the restart annotation
if daemonSet.Spec.Template.Annotations == nil {
daemonSet.Spec.Template.Annotations = make(map[string]string)
}
	daemonSet.Spec.Template.Annotations["kubectl.kubernetes.io/restartedAt"] = time.Now().Format(time.RFC3339)

	// Update the daemonset
	_, err = daemonSetsClient.Update(ctx, daemonSet, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to update daemonset: %w", err)
	}

	return nil
}

// Info returns the configured restart service information
func (s *WorkloadService) Info(ctx context.Context, req *workloadv1.InfoRequest) (*workloadv1.ServiceInfo, error) {
	slog.InfoContext(ctx, "Received info request")

	return &workloadv1.ServiceInfo{
		Enabled: s.restartEnabled,
		Type:    s.restartType,
		Name:    s.restartName,
	}, nil
}

// Restart performs a restart using the pre-configured workload settings
func (s *WorkloadService) Restart(ctx context.Context, req *workloadv1.SimpleRestartRequest) (*workloadv1.RestartResponse, error) {
	slog.InfoContext(ctx, "Received restart request for configured service")

	if !s.restartEnabled {
		return &workloadv1.RestartResponse{
			Success: false,
			Message: "Restart feature is not enabled",
		}, nil
	}

	if s.restartName == "" {
		return &workloadv1.RestartResponse{
			Success: false,
			Message: "Restart name is not configured",
		}, nil
	}

	// Call RestartWorkload with the configured values
	return s.RestartWorkload(ctx, &workloadv1.RestartRequest{
		Type:      s.restartType,
		Name:      s.restartName,
		Namespace: "", // Use service namespace
	})
}
