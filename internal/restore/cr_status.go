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
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/vmware-tanzu/velero/pkg/plugin/velero"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// CRDStatusRestoreItemAction is a restore item action plugin for Velero
type CRDStatusRestoreItemAction struct {
	Log logrus.FieldLogger
}

// AppliesTo returns information indicating that the CRDStatusRestoreItemAction should be run while restoring PVCs.
func (p *CRDStatusRestoreItemAction) AppliesTo() (velero.ResourceSelector, error) {
	// We apply to ALL resources, then find if they're applicable later
	return velero.ResourceSelector{}, nil
}

func isCustomResource(apiversion string) bool {
	apiGroups := strings.SplitN(apiversion, "/", 2)

	// If we have no `/` assume we're using something like v1
	if len(apiGroups) == 1 {
		return false
	}
	// Its a core API
	if apiGroups[0] == "v1" {
		return false
	}
	// If it has .k8s.io
	if strings.HasSuffix(apiGroups[0], ".k8s.io") {
		return false
	}

	// All else fails, assume we're working with a CR
	return true
}

// Execute modifies the PVC's spec to use the volumesnapshot object as the data source ensuring that the newly provisioned volume
// can be pre-populated with data from the volumesnapshot.
func (p *CRDStatusRestoreItemAction) Execute(input *velero.RestoreItemActionExecuteInput) (*velero.RestoreItemActionExecuteOutput, error) {

	res := input.Item.UnstructuredContent()
	origRes := input.ItemFromBackup.UnstructuredContent()
	name, _, _ := unstructured.NestedString(res, "metadata", "name")
	namespace, _, _ := unstructured.NestedString(res, "metadata", "namespace")
	kind, _, _ := unstructured.NestedString(res, "kind")
	apiversion, _, _ := unstructured.NestedString(res, "apiVersion")

	log := p.Log.WithFields(logrus.Fields{
		"name":       name,
		"namespace":  namespace,
		"kind":       kind,
		"apiversion": apiversion,
		"plugin":     "CRDStatusRestoreItemAction",
	})

	log.Infof("Starting CRDStatusRestoreItemAction")

	if !isCustomResource(apiversion) {
		log.Infof("Resource is not a custom Resource")
		return &velero.RestoreItemActionExecuteOutput{
			UpdatedItem: input.Item,
		}, nil
	}

	log.Infof("Checking if Resource had a status field")
	statusField, found, err := unstructured.NestedMap(origRes, "status")
	if err != nil {
		return &velero.RestoreItemActionExecuteOutput{
			UpdatedItem: input.Item,
		}, err
	}
	if found {
		log.Infof("Re-setting status field: %v", statusField)
		unstructured.SetNestedMap(res, statusField, "status")
	}

	log.Infof("Returning from CRDStatusRestoreItemAction")

	return &velero.RestoreItemActionExecuteOutput{
		UpdatedItem: &unstructured.Unstructured{Object: res},
	}, nil
}
