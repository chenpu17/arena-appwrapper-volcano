// Copyright 2024 The Kubeflow Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package training

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/kubeflow/arena/pkg/apis/config"
	"github.com/kubeflow/arena/pkg/apis/types"
	"github.com/kubeflow/arena/pkg/k8saccesser"
	appwrapperv1beta2 "github.com/kubeflow/arena/pkg/operators/appwrapper-operator/apis/appwrapper/v1beta2"
	"github.com/kubeflow/arena/pkg/operators/appwrapper-operator/client/clientset/versioned"
)

const (
	// AppWrapper labels for pods
	appWrapperLabelName = "workload.codeflare.dev/appwrapper"
)

// chiefPodSuffixPattern matches pod names ending with -0 (e.g., job-worker-0, job-master-0)
// but NOT job-10, job-20, etc. Requires a non-digit character before -0
var chiefPodSuffixPattern = regexp.MustCompile(`[^0-9]-0$`)

// AppWrapperJob represents an AppWrapper training job
type AppWrapperJob struct {
	*BasicJobInfo
	appwrapper   *appwrapperv1beta2.AppWrapper
	pods         []*corev1.Pod
	chiefPod     *corev1.Pod
	requestedGPU int64
	allocatedGPU int64
	trainerType  types.TrainingJobType
}

func (aj *AppWrapperJob) Name() string {
	return aj.name
}

func (aj *AppWrapperJob) Uid() string {
	return string(aj.appwrapper.UID)
}

// ChiefPod returns the chief pod of the job
func (aj *AppWrapperJob) ChiefPod() *corev1.Pod {
	return aj.chiefPod
}

func (aj *AppWrapperJob) Trainer() types.TrainingJobType {
	return aj.trainerType
}

// AllPods returns all pods of the training job
func (aj *AppWrapperJob) AllPods() []*corev1.Pod {
	return aj.pods
}

func (aj *AppWrapperJob) GetTrainJob() interface{} {
	return aj.appwrapper
}

func (aj *AppWrapperJob) GetLabels() map[string]string {
	return aj.appwrapper.Labels
}

// GetStatus returns the status of the Job
func (aj *AppWrapperJob) GetStatus() string {
	status := string(types.TrainingJobPending)
	defer log.Debugf("Get status of AppWrapperJob %s: %s", aj.appwrapper.Name, status)

	if aj.appwrapper.Name == "" {
		return status
	}

	// Map AppWrapper phases to training job status
	switch aj.appwrapper.Status.Phase {
	case appwrapperv1beta2.AppWrapperEmpty, appwrapperv1beta2.AppWrapperSuspended:
		status = string(types.TrainingJobQueuing)
	case appwrapperv1beta2.AppWrapperResuming:
		// Check conditions for more detailed status during resuming
		if aj.hasCondition(appwrapperv1beta2.AppWrapperConditionQuotaReserved, true) {
			if aj.hasCondition(appwrapperv1beta2.AppWrapperConditionResourcesDeployed, true) {
				status = string(types.TrainingJobPending) // Resources deployed, waiting for pods
			} else {
				status = string(types.TrainingJobPending) // Quota reserved, deploying resources
			}
		} else {
			status = string(types.TrainingJobQueuing) // Waiting for quota
		}
	case appwrapperv1beta2.AppWrapperRunning:
		// Check if unhealthy
		if aj.hasCondition(appwrapperv1beta2.AppWrapperConditionUnhealthy, true) {
			status = string(types.TrainingJobFailed)
		} else {
			status = string(types.TrainingJobRunning)
		}
	case appwrapperv1beta2.AppWrapperSucceeded:
		status = string(types.TrainingJobSucceeded)
	case appwrapperv1beta2.AppWrapperFailed:
		status = string(types.TrainingJobFailed)
	case appwrapperv1beta2.AppWrapperResetting, appwrapperv1beta2.AppWrapperSuspending, appwrapperv1beta2.AppWrapperTerminating:
		status = string(types.TrainingJobPending)
	default:
		status = string(types.TrainingJobPending)
	}

	return status
}

// hasCondition checks if the AppWrapper has a specific condition with the expected status
func (aj *AppWrapperJob) hasCondition(conditionType string, expectedStatus bool) bool {
	for _, cond := range aj.appwrapper.Status.Conditions {
		if string(cond.Type) == conditionType {
			if expectedStatus {
				return cond.Status == metav1.ConditionTrue
			}
			return cond.Status == metav1.ConditionFalse
		}
	}
	return false
}

// GetConditionMessage returns the message from a specific condition type
func (aj *AppWrapperJob) GetConditionMessage(conditionType string) string {
	for _, cond := range aj.appwrapper.Status.Conditions {
		if string(cond.Type) == conditionType {
			return cond.Message
		}
	}
	return ""
}

// StartTime returns the start time of the job
func (aj *AppWrapperJob) StartTime() *metav1.Time {
	return &aj.appwrapper.CreationTimestamp
}

// Age returns the age of the job
func (aj *AppWrapperJob) Age() time.Duration {
	if aj.appwrapper.CreationTimestamp.IsZero() {
		return 0
	}
	return metav1.Now().Sub(aj.appwrapper.CreationTimestamp.Time)
}

// Duration returns the training duration of the job
func (aj *AppWrapperJob) Duration() time.Duration {
	if aj.appwrapper.CreationTimestamp.IsZero() {
		return 0
	}

	// If job is completed or failed, calculate duration from conditions
	for _, cond := range aj.appwrapper.Status.Conditions {
		if cond.Type == appwrapperv1beta2.AppWrapperConditionPodsReady && cond.Status == metav1.ConditionTrue {
			if !cond.LastTransitionTime.IsZero() {
				return cond.LastTransitionTime.Sub(aj.appwrapper.CreationTimestamp.Time)
			}
		}
	}

	// If job is still running, return time since creation
	if aj.GetStatus() == string(types.TrainingJobRunning) || aj.GetStatus() == string(types.TrainingJobSucceeded) {
		return metav1.Now().Sub(aj.appwrapper.CreationTimestamp.Time)
	}

	return 0
}

// GetJobDashboards returns dashboard URLs for the job
func (aj *AppWrapperJob) GetJobDashboards(client *kubernetes.Clientset, namespace, arenaNamespace string) ([]string, error) {
	urls := []string{}

	dashboardURL, err := dashboard(client, namespace, "kubernetes-dashboard")
	if err != nil {
		log.Debugf("Get dashboard failed due to %v", err)
		dashboardURL, err = dashboard(client, arenaNamespace, "kubernetes-dashboard")
		if err != nil {
			log.Debugf("Get dashboard failed due to %v", err)
		}
	}

	if err != nil {
		log.Debugf("Get dashboard failed due to %v", err)
		dashboardURL, err = dashboard(client, "kube-system", "kubernetes-dashboard")
		if err != nil {
			log.Debugf("Get dashboard failed due to %v", err)
		}
	}

	if dashboardURL == "" {
		return urls, fmt.Errorf("no LOGVIEWER Installed")
	}

	if aj.chiefPod == nil || len(aj.chiefPod.Spec.Containers) == 0 {
		return urls, fmt.Errorf("appwrapper job is not ready")
	}

	url := fmt.Sprintf("%s/#!/log/%s/%s/%s?namespace=%s\n",
		dashboardURL,
		aj.chiefPod.Namespace,
		aj.chiefPod.Name,
		aj.chiefPod.Spec.Containers[0].Name,
		aj.chiefPod.Namespace)

	urls = append(urls, url)
	return urls, nil
}

// RequestedGPU returns the requested GPU count
func (aj *AppWrapperJob) RequestedGPU() int64 {
	if aj.requestedGPU > 0 {
		return aj.requestedGPU
	}
	requestGPUs := getRequestGPUsOfJobFromPodAnnotation(aj.pods)
	if requestGPUs > 0 {
		return requestGPUs
	}
	for _, pod := range aj.pods {
		aj.requestedGPU += gpuInPod(*pod)
	}
	return aj.requestedGPU
}

// AllocatedGPU returns the allocated GPU count
func (aj *AppWrapperJob) AllocatedGPU() int64 {
	if aj.allocatedGPU > 0 {
		return aj.allocatedGPU
	}
	for _, pod := range aj.pods {
		aj.allocatedGPU += gpuInActivePod(*pod)
	}
	return aj.allocatedGPU
}

// HostIPOfChief returns the host IP of the chief pod
func (aj *AppWrapperJob) HostIPOfChief() (hostIP string) {
	hostIP = "N/A"
	if aj.GetStatus() == string(types.TrainingJobRunning) && aj.chiefPod != nil {
		hostIP = aj.chiefPod.Status.HostIP
	}
	return hostIP
}

func (aj *AppWrapperJob) Namespace() string {
	return aj.appwrapper.Namespace
}

// GetPriorityClass returns the priority class name
func (aj *AppWrapperJob) GetPriorityClass() string {
	if aj.chiefPod != nil {
		return aj.chiefPod.Spec.PriorityClassName
	}
	return ""
}

// AppWrapperJobTrainer is the trainer for AppWrapper jobs
type AppWrapperJobTrainer struct {
	client           *kubernetes.Clientset
	appwrapperClient *versioned.Clientset
	trainerType      types.TrainingJobType
	enabled          bool
}

// NewAppWrapperJobTrainer creates a new AppWrapperJobTrainer
func NewAppWrapperJobTrainer() Trainer {
	enable := false
	appwrapperClient, err := versioned.NewForConfig(config.GetArenaConfiger().GetRestConfig())
	if err != nil {
		log.Debugf("AppWrapperJobTrainer client creation failed: %v", err)
	}

	_, err = config.GetArenaConfiger().GetAPIExtensionClientSet().ApiextensionsV1().CustomResourceDefinitions().Get(context.TODO(), k8saccesser.AppWrapperCRDName, metav1.GetOptions{})
	if err == nil {
		log.Debugf("AppWrapperJobTrainer is enabled")
		enable = true
	} else {
		log.Debugf("AppWrapperJobTrainer is disabled, reason: %v", err)
	}

	log.Debugf("Succeed to init AppWrapperJobTrainer")
	return &AppWrapperJobTrainer{
		appwrapperClient: appwrapperClient,
		client:           config.GetArenaConfiger().GetClientSet(),
		trainerType:      types.AppWrapperJob,
		enabled:          enable,
	}
}

// IsEnabled returns whether the trainer is enabled
func (at *AppWrapperJobTrainer) IsEnabled() bool {
	return at.enabled
}

// Type returns the trainer type
func (at *AppWrapperJobTrainer) Type() types.TrainingJobType {
	return at.trainerType
}

// IsSupported checks if the job is supported
func (at *AppWrapperJobTrainer) IsSupported(name, ns string) bool {
	if !at.enabled {
		return false
	}
	_, err := at.GetTrainingJob(name, ns)
	return err == nil
}

// GetTrainingJob retrieves a training job by name and namespace
func (at *AppWrapperJobTrainer) GetTrainingJob(name, namespace string) (TrainingJob, error) {
	appwrapper, err := at.appwrapperClient.WorkloadV1beta2().AppWrappers(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		// Convert "not found" error to types.ErrTrainingJobNotFound
		if strings.Contains(err.Error(), fmt.Sprintf(`"%v" not found`, name)) {
			return nil, types.ErrTrainingJobNotFound
		}
		return nil, err
	}

	if err := CheckJobIsOwnedByTrainer(appwrapper.Labels); err != nil {
		return nil, err
	}

	// Find pods associated with this AppWrapper
	allPods, err := k8saccesser.GetK8sResourceAccesser().ListPods(namespace, fmt.Sprintf("%s=%s", appWrapperLabelName, name), "", nil)
	if err != nil {
		return nil, err
	}

	pods, chiefPod := getPodsOfAppWrapperJob(at, appwrapper, allPods)

	return &AppWrapperJob{
		BasicJobInfo: &BasicJobInfo{
			resources: at.resources(name, namespace, pods),
			name:      name,
		},
		appwrapper:  appwrapper,
		chiefPod:    chiefPod,
		pods:        pods,
		trainerType: at.Type(),
	}, nil
}

func (at *AppWrapperJobTrainer) isChiefPod(appwrapper *appwrapperv1beta2.AppWrapper, item *corev1.Pod) bool {
	// For PyTorch jobs wrapped in AppWrapper, check for master label
	if val, ok := item.Labels["pytorch-replica-type"]; ok && val == "master" {
		return true
	}
	if val, ok := item.Labels["training.kubeflow.org/replica-type"]; ok && val == "master" {
		return true
	}
	// For Volcano jobs wrapped in AppWrapper, check for driver role
	if val, ok := item.Labels["volcano-role"]; ok && val == "driver" {
		return true
	}
	// For Volcano jobs, check for task index 0 (first worker)
	if val, ok := item.Annotations["volcano.sh/task-index"]; ok && val == "0" {
		return true
	}
	// Fallback: use regex to match pod names ending with -0 (e.g., job-worker-0, job-master-0)
	// The pattern requires a non-digit before -0 to avoid matching job-10, job-20, etc.
	if chiefPodSuffixPattern.MatchString(item.Name) {
		return true
	}
	return false
}

func (at *AppWrapperJobTrainer) isAppWrapperPod(name, ns string, pod *corev1.Pod) bool {
	if pod.Namespace != ns {
		return false
	}
	// Check if pod belongs to this AppWrapper
	if val, ok := pod.Labels[appWrapperLabelName]; ok && val == name {
		return true
	}
	// Also check for release label (used by Arena's Helm charts)
	if val, ok := pod.Labels["release"]; ok && val == name {
		if val2, ok := pod.Labels["app"]; ok && val2 == string(types.AppWrapperJob) {
			return true
		}
	}
	return false
}

func (at *AppWrapperJobTrainer) resources(name string, namespace string, pods []*corev1.Pod) []Resource {
	return podResources(pods)
}

// ListTrainingJobs lists all AppWrapper training jobs
func (at *AppWrapperJobTrainer) ListTrainingJobs(namespace string, allNamespace bool) ([]TrainingJob, error) {
	if allNamespace {
		namespace = metav1.NamespaceAll
	}

	trainingJobs := []TrainingJob{}
	jobLabels := GetTrainingJobLabels(at.Type())

	// List all AppWrappers
	appwrappers, err := at.appwrapperClient.WorkloadV1beta2().AppWrappers(namespace).List(metav1.ListOptions{
		LabelSelector: jobLabels,
	})
	if err != nil {
		return trainingJobs, err
	}

	pods, err := k8saccesser.GetK8sResourceAccesser().ListPods(namespace, fmt.Sprintf("app=%v", at.Type()), "", nil)
	if err != nil {
		return nil, err
	}

	for _, aw := range appwrappers.Items {
		awCopy := aw
		filterPods, chiefPod := getPodsOfAppWrapperJob(at, &awCopy, pods)
		trainingJobs = append(trainingJobs, &AppWrapperJob{
			BasicJobInfo: &BasicJobInfo{
				resources: podResources(filterPods),
				name:      aw.Name,
			},
			appwrapper:  &awCopy,
			chiefPod:    chiefPod,
			pods:        filterPods,
			trainerType: at.Type(),
		})
	}

	return trainingJobs, nil
}

// getPodsOfAppWrapperJob filters pods belonging to an AppWrapper job
func getPodsOfAppWrapperJob(at *AppWrapperJobTrainer, appwrapper *appwrapperv1beta2.AppWrapper, podList []*corev1.Pod) ([]*corev1.Pod, *corev1.Pod) {
	return getPodsOfTrainingJob(appwrapper.Name, appwrapper.Namespace, podList, at.isAppWrapperPod, func(pod *corev1.Pod) bool {
		return at.isChiefPod(appwrapper, pod)
	})
}
