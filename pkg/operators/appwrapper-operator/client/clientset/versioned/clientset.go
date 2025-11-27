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

package versioned

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"

	appwrapperv1beta2 "github.com/kubeflow/arena/pkg/operators/appwrapper-operator/apis/appwrapper/v1beta2"
)

// AppWrapperV1beta2Interface defines the interface for AppWrapper v1beta2 API
type AppWrapperV1beta2Interface interface {
	AppWrappers(namespace string) AppWrapperInterface
}

// AppWrapperInterface defines the interface for operating on AppWrapper resources
type AppWrapperInterface interface {
	Get(name string, options metav1.GetOptions) (*appwrapperv1beta2.AppWrapper, error)
	List(options metav1.ListOptions) (*appwrapperv1beta2.AppWrapperList, error)
	Create(appwrapper *appwrapperv1beta2.AppWrapper) (*appwrapperv1beta2.AppWrapper, error)
	Delete(name string, options *metav1.DeleteOptions) error
}

// Clientset is the client for AppWrapper resources
type Clientset struct {
	dynamicClient dynamic.Interface
}

// NewForConfig creates a new Clientset for the given config
func NewForConfig(config *rest.Config) (*Clientset, error) {
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return &Clientset{dynamicClient: dynamicClient}, nil
}

// NewForConfigOrDie creates a new Clientset for the given config and panics on error
func NewForConfigOrDie(config *rest.Config) *Clientset {
	client, err := NewForConfig(config)
	if err != nil {
		panic(err)
	}
	return client
}

// WorkloadV1beta2 returns the AppWrapperV1beta2Interface
func (c *Clientset) WorkloadV1beta2() AppWrapperV1beta2Interface {
	return &appWrapperV1beta2Client{dynamicClient: c.dynamicClient}
}

type appWrapperV1beta2Client struct {
	dynamicClient dynamic.Interface
}

func (c *appWrapperV1beta2Client) AppWrappers(namespace string) AppWrapperInterface {
	return &appWrappers{
		dynamicClient: c.dynamicClient,
		namespace:     namespace,
	}
}

type appWrappers struct {
	dynamicClient dynamic.Interface
	namespace     string
}

var appWrapperGVR = schema.GroupVersionResource{
	Group:    "workload.codeflare.dev",
	Version:  "v1beta2",
	Resource: "appwrappers",
}

func (a *appWrappers) Get(name string, options metav1.GetOptions) (*appwrapperv1beta2.AppWrapper, error) {
	unstructuredObj, err := a.dynamicClient.Resource(appWrapperGVR).Namespace(a.namespace).Get(context.TODO(), name, options)
	if err != nil {
		return nil, err
	}
	return convertToAppWrapper(unstructuredObj)
}

func (a *appWrappers) List(options metav1.ListOptions) (*appwrapperv1beta2.AppWrapperList, error) {
	unstructuredList, err := a.dynamicClient.Resource(appWrapperGVR).Namespace(a.namespace).List(context.TODO(), options)
	if err != nil {
		return nil, err
	}
	return convertToAppWrapperList(unstructuredList)
}

func (a *appWrappers) Create(appwrapper *appwrapperv1beta2.AppWrapper) (*appwrapperv1beta2.AppWrapper, error) {
	unstructuredObj, err := convertFromAppWrapper(appwrapper)
	if err != nil {
		return nil, err
	}
	created, err := a.dynamicClient.Resource(appWrapperGVR).Namespace(a.namespace).Create(context.TODO(), unstructuredObj, metav1.CreateOptions{})
	if err != nil {
		return nil, err
	}
	return convertToAppWrapper(created)
}

func (a *appWrappers) Delete(name string, options *metav1.DeleteOptions) error {
	return a.dynamicClient.Resource(appWrapperGVR).Namespace(a.namespace).Delete(context.TODO(), name, *options)
}

// convertToAppWrapper converts an unstructured object to AppWrapper
func convertToAppWrapper(obj *unstructured.Unstructured) (*appwrapperv1beta2.AppWrapper, error) {
	aw := &appwrapperv1beta2.AppWrapper{}
	aw.Name = obj.GetName()
	aw.Namespace = obj.GetNamespace()
	aw.UID = obj.GetUID()
	aw.Labels = obj.GetLabels()
	aw.Annotations = obj.GetAnnotations()
	aw.CreationTimestamp = obj.GetCreationTimestamp()

	// Parse spec
	spec, found, err := unstructured.NestedMap(obj.Object, "spec")
	if err != nil {
		return nil, err
	}
	if found {
		if suspend, ok, _ := unstructured.NestedBool(spec, "suspend"); ok {
			aw.Spec.Suspend = suspend
		}
		// Parse components
		if components, ok, _ := unstructured.NestedSlice(spec, "components"); ok {
			for range components {
				aw.Spec.Components = append(aw.Spec.Components, appwrapperv1beta2.AppWrapperComponent{})
			}
		}
	}

	// Parse status
	status, found, err := unstructured.NestedMap(obj.Object, "status")
	if err != nil {
		return nil, err
	}
	if found {
		if phase, ok, _ := unstructured.NestedString(status, "phase"); ok {
			aw.Status.Phase = appwrapperv1beta2.AppWrapperPhase(phase)
		}
		if retries, ok, _ := unstructured.NestedInt64(status, "resettingCount"); ok {
			aw.Status.Retries = int32(retries)
		}
	}

	return aw, nil
}

// convertToAppWrapperList converts an unstructured list to AppWrapperList
func convertToAppWrapperList(list *unstructured.UnstructuredList) (*appwrapperv1beta2.AppWrapperList, error) {
	awList := &appwrapperv1beta2.AppWrapperList{}
	for _, item := range list.Items {
		aw, err := convertToAppWrapper(&item)
		if err != nil {
			return nil, err
		}
		awList.Items = append(awList.Items, *aw)
	}
	return awList, nil
}

// convertFromAppWrapper converts AppWrapper to unstructured object
func convertFromAppWrapper(aw *appwrapperv1beta2.AppWrapper) (*unstructured.Unstructured, error) {
	obj := &unstructured.Unstructured{}
	obj.SetAPIVersion("workload.codeflare.dev/v1beta2")
	obj.SetKind("AppWrapper")
	obj.SetName(aw.Name)
	obj.SetNamespace(aw.Namespace)
	obj.SetLabels(aw.Labels)
	obj.SetAnnotations(aw.Annotations)

	// Set spec
	spec := map[string]interface{}{
		"suspend": aw.Spec.Suspend,
	}

	// Convert components
	components := make([]interface{}, 0, len(aw.Spec.Components))
	for _, comp := range aw.Spec.Components {
		compMap := map[string]interface{}{}
		if comp.Annotations != nil {
			compMap["annotations"] = comp.Annotations
		}
		if comp.Template.Raw != nil {
			compMap["template"] = comp.Template.Raw
		}
		components = append(components, compMap)
	}
	if len(components) > 0 {
		spec["components"] = components
	}

	if err := unstructured.SetNestedMap(obj.Object, spec, "spec"); err != nil {
		return nil, err
	}

	return obj, nil
}
