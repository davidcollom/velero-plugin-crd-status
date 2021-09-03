/*
Copyright 2019, 2020 the Velero contributors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package restore

import (
	"github.com/sirupsen/logrus"
	"github.com/vmware-tanzu/velero/pkg/plugin/velero"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

const (
	AnnToRemove = "kubectl.kubernetes.io/last-applied-configuration"
)

// ResourceLastKnownAnnotation is a restore item action plugin for Velero
type ResourceLastKnownAnnotation struct {
	Log logrus.FieldLogger
}

// AppliesTo returns information indicating that the ResourceLastKnownAnnotation should be run while restoring all resources.
func (p *ResourceLastKnownAnnotation) AppliesTo() (velero.ResourceSelector, error) {
	return velero.ResourceSelector{}, nil
}

// Execute removes the kubectl last-applied-configuration annotations to work around: https://feichashao.com/kubectl-apply-fail/
func (p *ResourceLastKnownAnnotation) Execute(input *velero.RestoreItemActionExecuteInput) (*velero.RestoreItemActionExecuteOutput, error) {

	res := input.Item.UnstructuredContent()
	name, _, _ := unstructured.NestedString(res, "metadata", "name")
	namespace, _, _ := unstructured.NestedString(res, "metadata", "namespace")
	kind, _, _ := unstructured.NestedString(res, "kind")

	log := p.Log.WithFields(logrus.Fields{
		"name":      name,
		"namespace": namespace,
		"kind":      kind,
		"plugin":    "ResourceLastKnownAnnotation",
	})

	log.Infof("Starting ResourceLastKnownAnnotation")

	annotations, found, err := unstructured.NestedStringMap(res, "metadata", "annotations")
	if err != nil {
		log.Error(err)
		return &velero.RestoreItemActionExecuteOutput{
			UpdatedItem: &unstructured.Unstructured{Object: res},
		}, err
	}
	// Not found, lets initialize annotations
	if !found {
		log.Infof("Item has no annotations, ignoring.")
		return &velero.RestoreItemActionExecuteOutput{
			UpdatedItem: &unstructured.Unstructured{Object: res},
		}, nil
	}

	_, ok := annotations[AnnToRemove]
	if ok {
		log.Infof("Removing %s annoation", AnnToRemove)
		delete(annotations, AnnToRemove)
	}

	log.Infof("About to set Annotations too (%v)", annotations)
	err = unstructured.SetNestedStringMap(res, annotations, "metadata", "annotations")
	if err != nil {
		log.Error(err)
		return &velero.RestoreItemActionExecuteOutput{
			UpdatedItem: &unstructured.Unstructured{Object: res},
		}, err
	}

	log.Infof("Returning from ResourceLastKnownAnnotation")

	return &velero.RestoreItemActionExecuteOutput{
		UpdatedItem: &unstructured.Unstructured{Object: res},
	}, nil
}
