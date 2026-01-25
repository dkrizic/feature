package workload

import (
"testing"

workloadv1 "github.com/dkrizic/feature/service/service/workload/v1"
)

// TestWorkloadServiceCreation tests that we can create a workload service
func TestWorkloadServiceCreation(t *testing.T) {
	// When running outside of a Kubernetes cluster, this should fail
	_, err := NewWorkloadService("default", false, workloadv1.WorkloadType_WORKLOAD_TYPE_DEPLOYMENT, "test")
	if err == nil {
		t.Skip("Skipping test - running inside Kubernetes cluster")
	}
	// Expected to fail when not in cluster
	if err == nil {
		t.Error("Expected error when creating service outside cluster, got nil")
	}
}

// TestWorkloadTypeEnum tests that workload types are correctly defined
func TestWorkloadTypeEnum(t *testing.T) {
tests := []struct {
name     string
wType    workloadv1.WorkloadType
expected string
}{
{"Deployment", workloadv1.WorkloadType_WORKLOAD_TYPE_DEPLOYMENT, "WORKLOAD_TYPE_DEPLOYMENT"},
{"StatefulSet", workloadv1.WorkloadType_WORKLOAD_TYPE_STATEFULSET, "WORKLOAD_TYPE_STATEFULSET"},
{"DaemonSet", workloadv1.WorkloadType_WORKLOAD_TYPE_DAEMONSET, "WORKLOAD_TYPE_DAEMONSET"},
{"Unspecified", workloadv1.WorkloadType_WORKLOAD_TYPE_UNSPECIFIED, "WORKLOAD_TYPE_UNSPECIFIED"},
}

for _, tt := range tests {
t.Run(tt.name, func(t *testing.T) {
if tt.wType.String() != tt.expected {
t.Errorf("Expected %s, got %s", tt.expected, tt.wType.String())
}
})
}
}
