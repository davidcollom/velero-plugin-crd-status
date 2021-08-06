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
	"encoding/json"
	"strings"

	"github.com/sirupsen/logrus"
	"go.etcd.io/etcd/error"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/vmware-tanzu/velero/pkg/plugin/velero"
)

const (
// AnnBindCompleted      = "pv.kubernetes.io/bind-completed"
// AnnBoundByController  = "pv.kubernetes.io/bound-by-controller"
// AnnStorageProvisioner = "volume.beta.kubernetes.io/storage-provisioner"
// AnnSelectedNode       = "volume.kubernetes.io/selected-node"
)

// newCRDStatusRestoreItemAction is a restore item action plugin for Velero
type CRDStatusRestoreItemAction struct {
	Log logrus.FieldLogger
}

// AppliesTo returns information indicating that the CRDStatusRestoreItemAction should be run while restoring PVCs.
func (p *CRDStatusRestoreItemAction) AppliesTo() (velero.ResourceSelector, error) {
	// We apply to ALL resources, then find if they're applicable later
	return velero.ResourceSelector{}, nil
}

// func resourceHasStatusSubResource() bool {
// 	return false
// }

func isCustomResource(item runtime.Unstructured) bool {
	apiversion, _, err := unstructured.NestedString(item.UnstructuredContent(), "apiversion")
	apiGroups := strings.SplitN(apiversion, "/", 2)

	// If we have no `/` assume we're using something like v1
	if len(apiGroups) == 1 { return false }
	// Its a core API
	if apiGroups[0] == "v1" { return false }
	// If it has .k8s.io
	if strings.HasSuffix(apiGroups[0], ".k8s.io") { return false}

	// All else fails, assume we're working with a CR
	return true
}

func hasStatusField(item runtime.Unstructured) bool {
	_, status_field, err := unstructured.NestedSlice(item.UnstructuredContent(), "status")
	return status_field
}

// Execute modifies the PVC's spec to use the volumesnapshot object as the data source ensuring that the newly provisioned volume
// can be pre-populated with data from the volumesnapshot.
func (p *CRDStatusRestoreItemAction) Execute(input *velero.RestoreItemActionExecuteInput) (*velero.RestoreItemActionExecuteOutput, error) {
	name, _, err := unstructured.NestedString(input.Item.UnstructuredContent(), "metadata", "name")
	namespace, _, err := unstructured.NestedString(input.Item.UnstructuredContent(), "metadata", "namespace")

	res := input.Item

	p.Log.Infof("Starting CRDStatusRestoreItemAction for Resource %s/%s", itemJSON["metadata"]["namespace"], itemJSON["metadata"]["name"])

	if !isCustomResource(input.Item) {
		return &velero.RestoreItemActionExecuteOutput{
			UpdatedItem: input.Item,
		}, nil
	}
	if hasStatusField(input.item){

	}

	p.Log.Infof("Returning from CRDStatusRestoreItemAction for PVC %s/%s", itemJSON["metadata"]["namespace"], itemJSON["metadata"]["name"]))

	return &velero.RestoreItemActionExecuteOutput{
		UpdatedItem: &unstructured.Unstructured{Object: pvcMap},
	}, nil
}
