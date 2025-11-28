// Copyright 2024 The Kubeflow Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License

package argsbuilder

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/kubeflow/arena/pkg/apis/types"
)

type SubmitAppWrapperJobArgsBuilder struct {
	args        *types.SubmitAppWrapperJobArgs
	argValues   map[string]interface{}
	subBuilders map[string]ArgsBuilder
}

func NewSubmitAppWrapperJobArgsBuilder(args *types.SubmitAppWrapperJobArgs) ArgsBuilder {
	args.TrainingType = types.AppWrapperJob
	s := &SubmitAppWrapperJobArgsBuilder{
		args:        args,
		argValues:   map[string]interface{}{},
		subBuilders: map[string]ArgsBuilder{},
	}
	s.AddSubBuilder(
		NewSubmitArgsBuilder(&s.args.CommonSubmitArgs),
		NewSubmitSyncCodeArgsBuilder(&s.args.SubmitSyncCodeArgs),
		NewSubmitTensorboardArgsBuilder(&s.args.SubmitTensorboardArgs),
	)
	return s
}

func (s *SubmitAppWrapperJobArgsBuilder) GetName() string {
	items := strings.Split(fmt.Sprintf("%v", reflect.TypeOf(*s)), ".")
	return items[len(items)-1]
}

func (s *SubmitAppWrapperJobArgsBuilder) AddSubBuilder(builders ...ArgsBuilder) ArgsBuilder {
	for _, b := range builders {
		s.subBuilders[b.GetName()] = b
	}
	return s
}

func (s *SubmitAppWrapperJobArgsBuilder) AddArgValue(key string, value interface{}) ArgsBuilder {
	for name := range s.subBuilders {
		s.subBuilders[name].AddArgValue(key, value)
	}
	s.argValues[key] = value
	return s
}

func (s *SubmitAppWrapperJobArgsBuilder) AddCommandFlags(command *cobra.Command) {
	for name := range s.subBuilders {
		s.subBuilders[name].AddCommandFlags(command)
	}

	var (
		runningTimeout   time.Duration
		ttlAfterFinished time.Duration
	)

	// Basic resource settings (inherited from PyTorch pattern)
	command.Flags().StringVar(&s.args.CleanPodPolicy, "clean-task-policy", "Running", "How to clean tasks after Training is done, support None, Running, All.")
	command.Flags().StringVar(&s.args.Cpu, "cpu", "", "the cpu resource to use for the training, like 1 for 1 core.")
	command.Flags().StringVar(&s.args.Memory, "memory", "", "the memory resource to use for the training, like 1Gi.")
	command.Flags().DurationVar(&runningTimeout, "running-timeout", runningTimeout, "Specifies the duration since startTime during which the job can remain active before it is terminated(e.g. '5s', '1m', '2h22m').")
	command.Flags().DurationVar(&ttlAfterFinished, "ttl-after-finished", ttlAfterFinished, "Defines the TTL for cleaning up finished jobs(e.g. '5s', '1m', '2h22m'). Defaults to infinite.")
	command.Flags().StringVar(&s.args.ShareMemory, "share-memory", "2Gi", "the shared memory of each replica to run the job, default 2Gi.")
	command.Flags().StringVar(&s.args.NprocPerNode, "nproc-per-node", "", "The number of processes per node, available values are \"auto\", \"cpu\", \"gpu\" and a number (e.g. 4).")

	// AppWrapper specific settings
	command.Flags().StringVar(&s.args.KueueQueueName, "kueue-queue", "", "The Kueue LocalQueue name for resource quota management.")
	command.Flags().Int32Var(&s.args.RetryLimit, "retry-limit", 3, "Maximum number of retries before marking as Failed.")
	command.Flags().StringVar(&s.args.AdmissionGracePeriod, "admission-grace-period", "1m", "Duration to wait for pods to be admitted (e.g. '1m', '30s').")
	command.Flags().StringVar(&s.args.WarmupGracePeriod, "warmup-grace-period", "5m", "Duration to wait for pods to become ready (e.g. '5m', '1m30s').")
	command.Flags().StringVar(&s.args.FailureGracePeriod, "failure-grace-period", "1m", "Duration to wait before treating a failure as permanent (e.g. '1m', '30s').")
	command.Flags().StringVar(&s.args.RetryPausePeriod, "retry-pause-period", "90s", "Duration to pause between retries (e.g. '90s', '2m').")
	command.Flags().StringVar(&s.args.SuccessTTL, "success-ttl", "", "Duration after which a successful AppWrapper is deleted (e.g. '1h', '24h'). If not set, not auto-deleted.")
	command.Flags().StringVar(&s.args.InnerJobType, "inner-type", "pytorch", "The type of job wrapped inside AppWrapper. Supports 'pytorch' and 'volcano'.")

	// Volcano Job specific settings
	command.Flags().Int32Var(&s.args.MinAvailable, "min-available", 0, "Minimum number of pods that must be available (Volcano). If not set, defaults to replicas.")
	command.Flags().StringVar(&s.args.SchedulerName, "scheduler-name", "volcano", "Scheduler to use for the job (Volcano).")
	command.Flags().StringVar(&s.args.TaskName, "task-name", "worker", "Name of the task in Volcano Job.")
	command.Flags().Int32Var(&s.args.MaxRetry, "max-retry", 10000, "Maximum number of retries for a task (Volcano).")
	command.Flags().Int32Var(&s.args.Replicas, "replicas", 1, "Number of replicas for Volcano Job tasks.")
	command.Flags().Int32Var(&s.args.MasterPort, "master-port", 23456, "Port for distributed training communication (Volcano).")

	// Network Topology settings (Volcano)
	command.Flags().StringVar(&s.args.NetworkTopologyMode, "network-topology-mode", "", "Network topology mode: 'hard' (must satisfy) or 'soft' (prefer).")
	command.Flags().Int32Var(&s.args.HighestTierAllowed, "highest-tier-allowed", 0, "Highest network topology tier allowed. Lower tier = lower latency.")

	// Partition Policy settings (Volcano)
	command.Flags().Int32Var(&s.args.TotalPartitions, "total-partitions", 0, "Total number of partitions for distributed tasks.")
	command.Flags().Int32Var(&s.args.PartitionSize, "partition-size", 0, "Number of pods per partition.")
	command.Flags().StringVar(&s.args.PartitionNetworkTopologyMode, "partition-topology-mode", "", "Network topology mode within partitions.")
	command.Flags().Int32Var(&s.args.PartitionHighestTierAllowed, "partition-highest-tier", 0, "Highest tier allowed within partitions.")

	// Hardware specific settings
	command.Flags().StringVar(&s.args.RingController, "ring-controller", "", "Ring controller label for hardware affinity (e.g. 'ascend-1980').")

	s.AddArgValue("running-timeout", &runningTimeout).
		AddArgValue("ttl-after-finished", &ttlAfterFinished)
}

func (s *SubmitAppWrapperJobArgsBuilder) PreBuild() error {
	for name := range s.subBuilders {
		if err := s.subBuilders[name].PreBuild(); err != nil {
			return err
		}
	}
	s.AddArgValue(ShareDataPrefix+"dataset", s.args.DataSet)
	return nil
}

func (s *SubmitAppWrapperJobArgsBuilder) Build() error {
	for name := range s.subBuilders {
		if err := s.subBuilders[name].Build(); err != nil {
			return err
		}
	}
	// For Volcano mode, sync Replicas to WorkerCount and update related values
	// This must be done AFTER sub-builders run, as they set envs["workers"],
	// PodGroupMinAvailable, and request-gpus based on WorkerCount
	if err := s.syncVolcanoWorkerCount(); err != nil {
		return err
	}
	if err := s.setRunPolicy(); err != nil {
		return err
	}
	if err := s.check(); err != nil {
		return err
	}
	if err := s.setAppWrapperAnnotations(); err != nil {
		return err
	}
	if err := s.addEnv(); err != nil {
		return err
	}
	return nil
}

// syncVolcanoWorkerCount synchronizes Replicas to WorkerCount for Volcano mode
// and updates all related values that were set by sub-builders using the old WorkerCount
func (s *SubmitAppWrapperJobArgsBuilder) syncVolcanoWorkerCount() error {
	if s.args.InnerJobType != "volcano" {
		return nil
	}

	// Sync Replicas to WorkerCount for consistent resource statistics
	s.args.WorkerCount = int(s.args.Replicas)

	// Update envs["workers"] which was set by SubmitArgsBuilder.setJobInfoToEnv()
	if s.args.Envs == nil {
		s.args.Envs = map[string]string{}
	}
	s.args.Envs["workers"] = strconv.Itoa(s.args.WorkerCount)

	// Update PodGroupMinAvailable if coscheduling is enabled
	// This was set by SubmitArgsBuilder.addPodGroupLabel()
	if s.args.Coscheduling {
		s.args.PodGroupMinAvailable = fmt.Sprintf("%v", s.args.WorkerCount)
	}

	// Update request-gpus annotation which was set by SubmitArgsBuilder.addRequestGPUsToAnnotation()
	if s.args.Annotations == nil {
		s.args.Annotations = map[string]string{}
	}
	s.args.Annotations[types.RequestGPUsOfJobAnnoKey] = fmt.Sprintf("%v", s.args.WorkerCount*s.args.GPUCount)

	return nil
}

func (s *SubmitAppWrapperJobArgsBuilder) setRunPolicy() error {
	// Get active deadline
	if rt, ok := s.argValues["running-timeout"]; ok {
		runningTimeout := rt.(*time.Duration)
		s.args.ActiveDeadlineSeconds = int64(runningTimeout.Seconds())
	}

	// Get ttlSecondsAfterFinished
	if ft, ok := s.argValues["ttl-after-finished"]; ok {
		ttlAfterFinished := ft.(*time.Duration)
		s.args.TTLSecondsAfterFinished = int32(ttlAfterFinished.Seconds())
	}
	return nil
}

func (s *SubmitAppWrapperJobArgsBuilder) check() error {
	if s.args.Image == "" {
		return fmt.Errorf("--image must be set")
	}

	// check clean-task-policy
	switch s.args.CleanPodPolicy {
	case "None", "Running", "All":
		log.Debugf("Supported cleanTaskPolicy: %s", s.args.CleanPodPolicy)
	default:
		return fmt.Errorf("unsupported cleanTaskPolicy %s", s.args.CleanPodPolicy)
	}

	if s.args.GPUCount < 0 {
		return fmt.Errorf("--gpus is invalid")
	}
	if s.args.Cpu != "" {
		_, err := resource.ParseQuantity(s.args.Cpu)
		if err != nil {
			return fmt.Errorf("--cpu is invalid")
		}
	}
	if s.args.Memory != "" {
		_, err := resource.ParseQuantity(s.args.Memory)
		if err != nil {
			return fmt.Errorf("--memory is invalid")
		}
	}
	if s.args.ActiveDeadlineSeconds < 0 {
		return fmt.Errorf("--running-timeout is invalid")
	}
	if s.args.TTLSecondsAfterFinished < 0 {
		return fmt.Errorf("--ttl-after-finished is invalid")
	}
	if s.args.ShareMemory != "" {
		_, err := resource.ParseQuantity(s.args.ShareMemory)
		if err != nil {
			return fmt.Errorf("--share-memory is invalid")
		}
	}

	// Check whether nprocPerNode is valid
	switch s.args.NprocPerNode {
	case "auto", "cpu", "gpu":
		log.Debugf("Supported nprocPerNode: %s", s.args.NprocPerNode)
	case "":
		log.Debugf("--nproc-per-node is not set")
	default:
		nprocPerNode, err := strconv.Atoi(s.args.NprocPerNode)
		if err != nil {
			return fmt.Errorf("--nproc-per-node is invalid")
		}
		log.Debugf("Supported nprocPerNode: %d", nprocPerNode)
	}

	// Check AppWrapper specific parameters
	if s.args.RetryLimit < 0 {
		return fmt.Errorf("--retry-limit must be >= 0")
	}

	// Validate duration formats
	durationParams := map[string]string{
		"admission-grace-period": s.args.AdmissionGracePeriod,
		"warmup-grace-period":    s.args.WarmupGracePeriod,
		"failure-grace-period":   s.args.FailureGracePeriod,
		"retry-pause-period":     s.args.RetryPausePeriod,
	}
	for paramName, paramValue := range durationParams {
		if paramValue != "" {
			_, err := time.ParseDuration(paramValue)
			if err != nil {
				return fmt.Errorf("--%s is invalid: %v", paramName, err)
			}
		}
	}

	if s.args.SuccessTTL != "" {
		_, err := time.ParseDuration(s.args.SuccessTTL)
		if err != nil {
			return fmt.Errorf("--success-ttl is invalid: %v", err)
		}
	}

	// Check inner job type
	switch s.args.InnerJobType {
	case "pytorch", "volcano":
		log.Debugf("Supported innerJobType: %s", s.args.InnerJobType)
	default:
		return fmt.Errorf("unsupported inner job type %s, supported types are 'pytorch' and 'volcano'", s.args.InnerJobType)
	}

	// Volcano-specific validations
	if s.args.InnerJobType == "volcano" {
		// Validate master port
		if s.args.MasterPort <= 0 || s.args.MasterPort > 65535 {
			return fmt.Errorf("--master-port must be between 1 and 65535, got %d", s.args.MasterPort)
		}

		// Validate network topology mode
		if s.args.NetworkTopologyMode != "" {
			switch s.args.NetworkTopologyMode {
			case "hard", "soft":
				log.Debugf("Supported networkTopologyMode: %s", s.args.NetworkTopologyMode)
			default:
				return fmt.Errorf("--network-topology-mode must be 'hard' or 'soft', got '%s'", s.args.NetworkTopologyMode)
			}
		}

		// Validate partition network topology mode
		if s.args.PartitionNetworkTopologyMode != "" {
			switch s.args.PartitionNetworkTopologyMode {
			case "hard", "soft":
				log.Debugf("Supported partitionNetworkTopologyMode: %s", s.args.PartitionNetworkTopologyMode)
			default:
				return fmt.Errorf("--partition-topology-mode must be 'hard' or 'soft', got '%s'", s.args.PartitionNetworkTopologyMode)
			}
		}

		// Validate partition policy consistency
		if s.args.TotalPartitions > 0 && s.args.PartitionSize <= 0 {
			return fmt.Errorf("--partition-size must be specified when --total-partitions is set")
		}
		if s.args.PartitionSize > 0 && s.args.TotalPartitions <= 0 {
			return fmt.Errorf("--total-partitions must be specified when --partition-size is set")
		}
	}

	return nil
}

// setAppWrapperAnnotations adds AppWrapper-specific annotations
func (s *SubmitAppWrapperJobArgsBuilder) setAppWrapperAnnotations() error {
	if s.args.Annotations == nil {
		s.args.Annotations = map[string]string{}
	}

	// Note: kueueQueueName is passed to Helm as a separate value (.Values.kueueQueueName)
	// and the Helm template only adds kueue.x-k8s.io/queue-name label at the AppWrapper level.
	// We should NOT add it to s.args.Labels here, as that would cause it to be propagated
	// to all child resources (Volcano Job, Pod templates) via .Values.labels.

	// Add AppWrapper fault tolerance configurations as annotations
	annotationPrefix := "workload.codeflare.dev.appwrapper/"

	if s.args.RetryLimit > 0 {
		s.args.Annotations[annotationPrefix+"retryLimit"] = fmt.Sprintf("%d", s.args.RetryLimit)
	}
	if s.args.AdmissionGracePeriod != "" {
		s.args.Annotations[annotationPrefix+"admissionGracePeriodDuration"] = s.args.AdmissionGracePeriod
	}
	if s.args.WarmupGracePeriod != "" {
		s.args.Annotations[annotationPrefix+"warmupGracePeriodDuration"] = s.args.WarmupGracePeriod
	}
	if s.args.FailureGracePeriod != "" {
		s.args.Annotations[annotationPrefix+"failureGracePeriodDuration"] = s.args.FailureGracePeriod
	}
	if s.args.RetryPausePeriod != "" {
		s.args.Annotations[annotationPrefix+"retryPausePeriodDuration"] = s.args.RetryPausePeriod
	}
	if s.args.SuccessTTL != "" {
		s.args.Annotations[annotationPrefix+"successTTLDuration"] = s.args.SuccessTTL
	}

	return nil
}

func (s *SubmitAppWrapperJobArgsBuilder) addEnv() error {
	if s.args.Envs == nil {
		s.args.Envs = map[string]string{}
	}

	// Only set MASTER_ADDR for PyTorchJob mode
	// For Volcano mode, MASTER_ADDR is set in the Helm template with correct DNS name
	if s.args.EnableRDMA && s.args.InnerJobType != "volcano" {
		s.args.Envs["MASTER_ADDR"] = fmt.Sprintf("%v-master-0", s.args.Name)
	}

	if s.args.NprocPerNode != "" {
		s.args.Envs["PET_NPROC_PER_NODE"] = s.args.NprocPerNode
	}

	return nil
}
