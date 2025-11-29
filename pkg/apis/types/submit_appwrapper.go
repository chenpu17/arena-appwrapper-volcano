// Copyright 2024 The Kubeflow Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package types

// SubmitAppWrapperJobArgs defines the arguments for submitting an AppWrapper job
type SubmitAppWrapperJobArgs struct {
	Cpu    string `yaml:"cpu"`    // --cpu
	Memory string `yaml:"memory"` // --memory

	// for common args
	CommonSubmitArgs `yaml:",inline"`

	// for tensorboard
	SubmitTensorboardArgs `yaml:",inline"`

	// for sync up source code
	SubmitSyncCodeArgs `yaml:",inline"`

	// CleanPodPolicy defines how to clean tasks after Training is done
	// Supported values: None, Running, All
	CleanPodPolicy string `yaml:"cleanPodPolicy"`

	// ActiveDeadlineSeconds specifies the duration (in seconds) since startTime
	// during which the job can remain active before it is terminated
	ActiveDeadlineSeconds int64 `yaml:"activeDeadlineSeconds,omitempty"`

	// TTLSecondsAfterFinished defines the TTL for cleaning up finished jobs
	TTLSecondsAfterFinished int32 `yaml:"ttlSecondsAfterFinished,omitempty"`

	// ShareMemory specifies the shared memory size for each replica
	ShareMemory string `yaml:"shareMemory"`

	// NprocPerNode is the number of processes per node
	// Supported values: auto, cpu, gpu, or a number
	NprocPerNode string `yaml:"nprocPerNode,omitempty"`

	// ========== AppWrapper specific parameters ==========

	// KueueQueueName specifies the Kueue LocalQueue name for the job
	// This is used to submit jobs to Kueue for resource quota management
	KueueQueueName string `yaml:"kueueQueueName,omitempty"`

	// RetryLimit specifies the maximum number of times the AppWrapper will be reset
	// after a failure before being marked as Failed. Default is 3.
	RetryLimit int32 `yaml:"retryLimit,omitempty"`

	// AdmissionGracePeriod specifies the duration to wait for pods to be admitted
	// Format: "1m", "30s", etc. Default is 1 minute.
	AdmissionGracePeriod string `yaml:"admissionGracePeriod,omitempty"`

	// WarmupGracePeriod specifies the duration to wait for pods to become ready
	// Format: "5m", "1m30s", etc. Default is 5 minutes.
	WarmupGracePeriod string `yaml:"warmupGracePeriod,omitempty"`

	// FailureGracePeriod specifies the duration to wait before treating a failure as permanent
	// Format: "1m", "30s", etc. Default is 1 minute.
	FailureGracePeriod string `yaml:"failureGracePeriod,omitempty"`

	// RetryPausePeriod specifies the duration to pause between retries
	// Format: "90s", "2m", etc. Default is 90 seconds.
	RetryPausePeriod string `yaml:"retryPausePeriod,omitempty"`

	// SuccessTTL specifies the duration after which a successful AppWrapper is deleted
	// Format: "1h", "24h", etc. If not set, the AppWrapper will not be auto-deleted.
	SuccessTTL string `yaml:"successTTL,omitempty"`

	// InnerJobType specifies the type of job wrapped inside AppWrapper
	// Supported values: "pytorch", "volcano"
	InnerJobType string `yaml:"innerJobType,omitempty"`

	// ========== Volcano Job specific parameters ==========

	// MinAvailable specifies the minimum number of pods that must be available
	// for the job to be considered ready. Used by Volcano scheduler.
	MinAvailable int32 `yaml:"minAvailable,omitempty"`

	// Note: SchedulerName is inherited from CommonSubmitArgs (--scheduler flag)

	// TaskName specifies the name of the task in Volcano Job
	TaskName string `yaml:"taskName,omitempty"`

	// MaxRetry specifies the maximum number of retries for a task
	MaxRetry int32 `yaml:"maxRetry,omitempty"`

	// Replicas specifies the number of replicas for Volcano Job tasks
	Replicas int32 `yaml:"replicas,omitempty"`

	// MasterPort specifies the port for distributed training communication
	// Default is 23456
	MasterPort int32 `yaml:"masterPort,omitempty"`

	// UseSvcPlugin enables Volcano's svc plugin for automatic Headless Service creation
	// Requires Volcano >= 1.8. Set to false to use manual Headless Service (fallback for older versions)
	// Default is true
	UseSvcPlugin *bool `yaml:"useSvcPlugin,omitempty"`

	// ========== Network Topology parameters (Volcano) ==========

	// NetworkTopologyMode specifies the network topology mode
	// Supported values: "hard" (must satisfy), "soft" (prefer)
	NetworkTopologyMode string `yaml:"networkTopologyMode,omitempty"`

	// HighestTierAllowed specifies the highest network topology tier allowed
	// Lower tier means lower network latency between nodes
	HighestTierAllowed int32 `yaml:"highestTierAllowed,omitempty"`

	// ========== Partition Policy parameters (Volcano) ==========

	// TotalPartitions specifies the total number of partitions for distributed tasks
	TotalPartitions int32 `yaml:"totalPartitions,omitempty"`

	// PartitionSize specifies the number of pods per partition
	PartitionSize int32 `yaml:"partitionSize,omitempty"`

	// PartitionNetworkTopologyMode specifies the network topology mode within partitions
	PartitionNetworkTopologyMode string `yaml:"partitionNetworkTopologyMode,omitempty"`

	// PartitionHighestTierAllowed specifies the highest tier allowed within partitions
	PartitionHighestTierAllowed int32 `yaml:"partitionHighestTierAllowed,omitempty"`

	// ========== Hardware specific parameters ==========

	// RingController specifies the ring controller label for hardware affinity
	// e.g., "ascend-1980" for Huawei Ascend NPU
	RingController string `yaml:"ringController,omitempty"`
}
