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

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/vmware-tanzu/velero/pkg/plugin/velero"
)

const (
	AnnPreviousTimestamp = "jetstack.io/original-timestamp"
)

type ResourceCreationTimestamp struct {
	Log logrus.FieldLogger
}

// AppliesTo returns information indicating that the ResourceCreationTimestamp should be run while restoring a Resource.
func (p *ResourceCreationTimestamp) AppliesTo() (velero.ResourceSelector, error) {
	return velero.ResourceSelector{}, nil
}

// Execute takes the metadata.creationTimestamp from the backup and adds to a custom annoation (ensuring that it is NOT overridden)
func (p *ResourceCreationTimestamp) Execute(input *velero.RestoreItemActionExecuteInput) (*velero.RestoreItemActionExecuteOutput, error) {

	res := input.Item.UnstructuredContent()
	origRes := input.ItemFromBackup.UnstructuredContent()
	name, _, _ := unstructured.NestedString(res, "metadata", "name")
	namespace, _, _ := unstructured.NestedString(res, "metadata", "namespace")
	kind, _, _ := unstructured.NestedString(res, "kind")

	log := p.Log.WithFields(logrus.Fields{
		"name":      name,
		"namespace": namespace,
		"kind":      kind,
		"plugin":    "ResourceCreationTimestamp",
	})

	log.Infof("Starting ResourceCreationTimestamp")

	annotations, found, err := unstructured.NestedStringMap(res, "metadata", "annotations")
	if err != nil {
		log.Error(err)
		return &velero.RestoreItemActionExecuteOutput{
			UpdatedItem: &unstructured.Unstructured{Object: res},
		}, err
	}
	// Not found, lets initialize annotations
	if !found {
		log.Infof("Annotations not found, initializing")
		annotations = make(map[string]string)
	}

	_, ok := annotations[AnnPreviousTimestamp]
	if ok {
		log.Infof("Skipping resource due to annoation already existing.")
		return &velero.RestoreItemActionExecuteOutput{
			UpdatedItem: &unstructured.Unstructured{Object: res},
		}, nil
	}

	log.Infof("Looking for original creationTimestamp")
	originalCreationTimestamp, found, err := unstructured.NestedString(origRes, "metadata", "creationTimestamp")
	if err != nil {
		log.Error(err)
		return &velero.RestoreItemActionExecuteOutput{
			UpdatedItem: &unstructured.Unstructured{Object: res},
		}, err
	}
	if found {
		log.Infof("Adding Annoation %s", AnnPreviousTimestamp)
		annotations[AnnPreviousTimestamp] = originalCreationTimestamp
	}
	log.Infof("About to set Annotations too (%v)", annotations)
	err = unstructured.SetNestedStringMap(res, annotations, "metadata", "annotations")
	if err != nil {
		log.Error(err)
		return &velero.RestoreItemActionExecuteOutput{
			UpdatedItem: &unstructured.Unstructured{Object: res},
		}, err
	}

	log.Infof("Returning from ResourceCreationTimestamp")

	return &velero.RestoreItemActionExecuteOutput{
		UpdatedItem: &unstructured.Unstructured{Object: res},
	}, nil
}
