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

package training

import (
	"fmt"
	"strings"

	"github.com/kubeflow/arena/pkg/apis/types"
	"github.com/kubeflow/arena/pkg/argsbuilder"
)

type AppWrapperJobBuilder struct {
	args      *types.SubmitAppWrapperJobArgs
	argValues map[string]interface{}
	argsbuilder.ArgsBuilder
}

func NewAppWrapperJobBuilder() *AppWrapperJobBuilder {
	args := &types.SubmitAppWrapperJobArgs{
		CleanPodPolicy:        "Running",
		CommonSubmitArgs:      DefaultCommonSubmitArgs,
		SubmitTensorboardArgs: DefaultSubmitTensorboardArgs,
		RetryLimit:            3,
		AdmissionGracePeriod:  "1m",
		WarmupGracePeriod:     "5m",
		FailureGracePeriod:    "1m",
		RetryPausePeriod:      "90s",
		InnerJobType:          "pytorch",
	}
	return &AppWrapperJobBuilder{
		args:        args,
		argValues:   map[string]interface{}{},
		ArgsBuilder: argsbuilder.NewSubmitAppWrapperJobArgsBuilder(args),
	}
}

// Name is used to set job name,match option --name
func (b *AppWrapperJobBuilder) Name(name string) *AppWrapperJobBuilder {
	if name != "" {
		b.args.Name = name
	}
	return b
}

// Shell is used to set bash or sh
func (b *AppWrapperJobBuilder) Shell(shell string) *AppWrapperJobBuilder {
	if shell != "" {
		b.args.Shell = shell
	}
	return b
}

// Command is used to set job command
func (b *AppWrapperJobBuilder) Command(args []string) *AppWrapperJobBuilder {
	if b.args.Command == "" {
		b.args.Command = strings.Join(args, " ")
	}
	return b
}

// WorkingDir is used to set working directory of job containers,default is '/root'
// match option --working-dir
func (b *AppWrapperJobBuilder) WorkingDir(dir string) *AppWrapperJobBuilder {
	if dir != "" {
		b.args.WorkingDir = dir
	}
	return b
}

// Envs is used to set env of job containers,match option --env
func (b *AppWrapperJobBuilder) Envs(envs map[string]string) *AppWrapperJobBuilder {
	if len(envs) != 0 {
		envSlice := []string{}
		for key, value := range envs {
			envSlice = append(envSlice, fmt.Sprintf("%v=%v", key, value))
		}
		b.argValues["env"] = &envSlice
	}
	return b
}

// GPUCount is used to set count of gpu for the job,match the option --gpus
func (b *AppWrapperJobBuilder) GPUCount(count int) *AppWrapperJobBuilder {
	if count > 0 {
		b.args.GPUCount = count
	}
	return b
}

// Devices is used to set chip vendors and count that used for resources
func (b *AppWrapperJobBuilder) Devices(devices map[string]string) *AppWrapperJobBuilder {
	if len(devices) != 0 {
		devicesSlice := []string{}
		for key, value := range devices {
			devicesSlice = append(devicesSlice, fmt.Sprintf("%v=%v", key, value))
		}
		b.argValues["device"] = &devicesSlice
	}
	return b
}

// Image is used to set job image,match the option --image
func (b *AppWrapperJobBuilder) Image(image string) *AppWrapperJobBuilder {
	if image != "" {
		b.args.Image = image
	}
	return b
}

// Tolerations is used to set tolerations for tolerate nodes,match option --toleration
func (b *AppWrapperJobBuilder) Tolerations(tolerations []string) *AppWrapperJobBuilder {
	b.argValues["toleration"] = &tolerations
	return b
}

// ConfigFiles is used to mapping config files form local to job containers
func (b *AppWrapperJobBuilder) ConfigFiles(files map[string]string) *AppWrapperJobBuilder {
	if len(files) != 0 {
		filesSlice := []string{}
		for localPath, containerPath := range files {
			filesSlice = append(filesSlice, fmt.Sprintf("%v:%v", localPath, containerPath))
		}
		b.argValues["config-file"] = &filesSlice
	}
	return b
}

// NodeSelectors is used to set node selectors for scheduling job
func (b *AppWrapperJobBuilder) NodeSelectors(selectors map[string]string) *AppWrapperJobBuilder {
	if len(selectors) != 0 {
		selectorsSlice := []string{}
		for key, value := range selectors {
			selectorsSlice = append(selectorsSlice, fmt.Sprintf("%v=%v", key, value))
		}
		b.argValues["selector"] = &selectorsSlice
	}
	return b
}

// Annotations is used to add annotations for job pods
func (b *AppWrapperJobBuilder) Annotations(annotations map[string]string) *AppWrapperJobBuilder {
	if len(annotations) != 0 {
		s := []string{}
		for key, value := range annotations {
			s = append(s, fmt.Sprintf("%v=%v", key, value))
		}
		b.argValues["annotation"] = &s
	}
	return b
}

// Labels is used to add labels for job
func (b *AppWrapperJobBuilder) Labels(labels map[string]string) *AppWrapperJobBuilder {
	if len(labels) != 0 {
		s := []string{}
		for key, value := range labels {
			s = append(s, fmt.Sprintf("%v=%v", key, value))
		}
		b.argValues["label"] = &s
	}
	return b
}

// Datas is used to mount k8s pvc to job pods,match option --data
func (b *AppWrapperJobBuilder) Datas(volumes map[string]string) *AppWrapperJobBuilder {
	if len(volumes) != 0 {
		s := []string{}
		for key, value := range volumes {
			s = append(s, fmt.Sprintf("%v:%v", key, value))
		}
		b.argValues["data"] = &s
	}
	return b
}

// DataDirs is used to mount host files to job containers
func (b *AppWrapperJobBuilder) DataDirs(volumes map[string]string) *AppWrapperJobBuilder {
	if len(volumes) != 0 {
		s := []string{}
		for key, value := range volumes {
			s = append(s, fmt.Sprintf("%v:%v", key, value))
		}
		b.argValues["data-dir"] = &s
	}
	return b
}

// LogDir is used to set log directory
func (b *AppWrapperJobBuilder) LogDir(dir string) *AppWrapperJobBuilder {
	if dir != "" {
		b.args.TrainingLogdir = dir
	}
	return b
}

// Priority sets the priority
func (b *AppWrapperJobBuilder) Priority(priority string) *AppWrapperJobBuilder {
	if priority != "" {
		b.args.PriorityClassName = priority
	}
	return b
}

// EnableRDMA is used to enabled rdma
func (b *AppWrapperJobBuilder) EnableRDMA() *AppWrapperJobBuilder {
	b.args.EnableRDMA = true
	return b
}

// SyncImage is used to set syncing image
func (b *AppWrapperJobBuilder) SyncImage(image string) *AppWrapperJobBuilder {
	if image != "" {
		b.args.SyncImage = image
	}
	return b
}

// SyncMode is used to set syncing mode
func (b *AppWrapperJobBuilder) SyncMode(mode string) *AppWrapperJobBuilder {
	if mode != "" {
		b.args.SyncMode = mode
	}
	return b
}

// SyncSource is used to set syncing source
func (b *AppWrapperJobBuilder) SyncSource(source string) *AppWrapperJobBuilder {
	if source != "" {
		b.args.SyncSource = source
	}
	return b
}

// EnableTensorboard is used to enable tensorboard
func (b *AppWrapperJobBuilder) EnableTensorboard() *AppWrapperJobBuilder {
	b.args.UseTensorboard = true
	return b
}

// TensorboardImage is used to set tensorboard image
func (b *AppWrapperJobBuilder) TensorboardImage(image string) *AppWrapperJobBuilder {
	if image != "" {
		b.args.TensorboardImage = image
	}
	return b
}

// ImagePullSecrets is used to set image pull secrets
func (b *AppWrapperJobBuilder) ImagePullSecrets(secrets []string) *AppWrapperJobBuilder {
	if secrets != nil {
		b.argValues["image-pull-secret"] = &secrets
	}
	return b
}

// CleanPodPolicy is used to set cleaning pod policy
func (b *AppWrapperJobBuilder) CleanPodPolicy(policy string) *AppWrapperJobBuilder {
	if policy != "" {
		b.args.CleanPodPolicy = policy
	}
	return b
}

// WorkerCount is used to set count of worker
func (b *AppWrapperJobBuilder) WorkerCount(count int) *AppWrapperJobBuilder {
	if count > 0 {
		b.args.WorkerCount = count
	}
	return b
}

// CPU assign cpu limits
func (b *AppWrapperJobBuilder) CPU(cpu string) *AppWrapperJobBuilder {
	if cpu != "" {
		b.args.Cpu = cpu
	}
	return b
}

// Memory assign memory limits
func (b *AppWrapperJobBuilder) Memory(memory string) *AppWrapperJobBuilder {
	if memory != "" {
		b.args.Memory = memory
	}
	return b
}

// ActiveDeadlineSeconds sets running timeout
func (b *AppWrapperJobBuilder) ActiveDeadlineSeconds(act int64) *AppWrapperJobBuilder {
	if act > 0 {
		b.args.ActiveDeadlineSeconds = act
	}
	return b
}

// TTLSecondsAfterFinished sets TTL after finished
func (b *AppWrapperJobBuilder) TTLSecondsAfterFinished(ttl int32) *AppWrapperJobBuilder {
	if ttl > 0 {
		b.args.TTLSecondsAfterFinished = ttl
	}
	return b
}

// ShareMemory sets shared memory size
func (b *AppWrapperJobBuilder) ShareMemory(shm string) *AppWrapperJobBuilder {
	if shm != "" {
		b.args.ShareMemory = shm
	}
	return b
}

// KueueQueueName sets the Kueue LocalQueue name
func (b *AppWrapperJobBuilder) KueueQueueName(queue string) *AppWrapperJobBuilder {
	if queue != "" {
		b.args.KueueQueueName = queue
	}
	return b
}

// RetryLimit sets the maximum number of retries
func (b *AppWrapperJobBuilder) RetryLimit(limit int32) *AppWrapperJobBuilder {
	if limit >= 0 {
		b.args.RetryLimit = limit
	}
	return b
}

// AdmissionGracePeriod sets the admission grace period
func (b *AppWrapperJobBuilder) AdmissionGracePeriod(period string) *AppWrapperJobBuilder {
	if period != "" {
		b.args.AdmissionGracePeriod = period
	}
	return b
}

// WarmupGracePeriod sets the warmup grace period
func (b *AppWrapperJobBuilder) WarmupGracePeriod(period string) *AppWrapperJobBuilder {
	if period != "" {
		b.args.WarmupGracePeriod = period
	}
	return b
}

// FailureGracePeriod sets the failure grace period
func (b *AppWrapperJobBuilder) FailureGracePeriod(period string) *AppWrapperJobBuilder {
	if period != "" {
		b.args.FailureGracePeriod = period
	}
	return b
}

// RetryPausePeriod sets the retry pause period
func (b *AppWrapperJobBuilder) RetryPausePeriod(period string) *AppWrapperJobBuilder {
	if period != "" {
		b.args.RetryPausePeriod = period
	}
	return b
}

// SuccessTTL sets the success TTL duration
func (b *AppWrapperJobBuilder) SuccessTTL(ttl string) *AppWrapperJobBuilder {
	if ttl != "" {
		b.args.SuccessTTL = ttl
	}
	return b
}

// InnerJobType sets the inner job type ("pytorch" or "volcano")
func (b *AppWrapperJobBuilder) InnerJobType(jobType string) *AppWrapperJobBuilder {
	if jobType != "" {
		b.args.InnerJobType = jobType
	}
	return b
}

// ========== Volcano Job specific methods ==========

// MinAvailable sets the minimum available pods for gang scheduling (Volcano)
func (b *AppWrapperJobBuilder) MinAvailable(minAvailable int32) *AppWrapperJobBuilder {
	if minAvailable > 0 {
		b.args.MinAvailable = minAvailable
	}
	return b
}

// SchedulerName sets the scheduler name for Volcano Job
func (b *AppWrapperJobBuilder) SchedulerName(name string) *AppWrapperJobBuilder {
	if name != "" {
		b.args.SchedulerName = name
	}
	return b
}

// TaskName sets the task name in Volcano Job
func (b *AppWrapperJobBuilder) TaskName(name string) *AppWrapperJobBuilder {
	if name != "" {
		b.args.TaskName = name
	}
	return b
}

// MaxRetry sets the maximum retry count for task failures
func (b *AppWrapperJobBuilder) MaxRetry(maxRetry int32) *AppWrapperJobBuilder {
	if maxRetry >= 0 {
		b.args.MaxRetry = maxRetry
	}
	return b
}

// Replicas sets the number of replicas for Volcano Job tasks
func (b *AppWrapperJobBuilder) Replicas(replicas int32) *AppWrapperJobBuilder {
	if replicas > 0 {
		b.args.Replicas = replicas
	}
	return b
}

// NetworkTopologyMode sets the network topology mode ("hard" or "soft")
func (b *AppWrapperJobBuilder) NetworkTopologyMode(mode string) *AppWrapperJobBuilder {
	if mode != "" {
		b.args.NetworkTopologyMode = mode
	}
	return b
}

// HighestTierAllowed sets the highest network topology tier allowed
func (b *AppWrapperJobBuilder) HighestTierAllowed(tier int32) *AppWrapperJobBuilder {
	if tier > 0 {
		b.args.HighestTierAllowed = tier
	}
	return b
}

// TotalPartitions sets the total number of partitions for distributed tasks
func (b *AppWrapperJobBuilder) TotalPartitions(partitions int32) *AppWrapperJobBuilder {
	if partitions > 0 {
		b.args.TotalPartitions = partitions
	}
	return b
}

// PartitionSize sets the number of pods per partition
func (b *AppWrapperJobBuilder) PartitionSize(size int32) *AppWrapperJobBuilder {
	if size > 0 {
		b.args.PartitionSize = size
	}
	return b
}

// PartitionNetworkTopologyMode sets the network topology mode within partitions
func (b *AppWrapperJobBuilder) PartitionNetworkTopologyMode(mode string) *AppWrapperJobBuilder {
	if mode != "" {
		b.args.PartitionNetworkTopologyMode = mode
	}
	return b
}

// PartitionHighestTierAllowed sets the highest tier allowed within partitions
func (b *AppWrapperJobBuilder) PartitionHighestTierAllowed(tier int32) *AppWrapperJobBuilder {
	if tier > 0 {
		b.args.PartitionHighestTierAllowed = tier
	}
	return b
}

// RingController sets the ring controller label for hardware affinity
func (b *AppWrapperJobBuilder) RingController(controller string) *AppWrapperJobBuilder {
	if controller != "" {
		b.args.RingController = controller
	}
	return b
}

// Build is used to build the job
func (b *AppWrapperJobBuilder) Build() (*Job, error) {
	for key, value := range b.argValues {
		b.AddArgValue(key, value)
	}
	if err := b.PreBuild(); err != nil {
		return nil, err
	}
	if err := b.ArgsBuilder.Build(); err != nil {
		return nil, err
	}
	return NewJob(b.args.Name, types.AppWrapperJob, b.args), nil
}
