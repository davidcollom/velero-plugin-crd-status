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
	"reflect"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/vmware-tanzu/velero/pkg/plugin/velero"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestIsCustomResource(t *testing.T) {
	tests := []struct {
		name       string
		apiVersion string
		want       bool
	}{
		{name: "core API version", apiVersion: "v1", want: false},
		{name: "Kubernetes API group suffix", apiVersion: "example.k8s.io/v1", want: false},
		{name: "version without group", apiVersion: "v1alpha1", want: false},
		{name: "group not in exclusion list", apiVersion: "apps/v1", want: true},
		{name: "custom API group", apiVersion: "example.io/v1alpha1", want: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := isCustomResource(test.apiVersion); got != test.want {
				t.Errorf("isCustomResource(%q) = %t, want %t", test.apiVersion, got, test.want)
			}
		})
	}
}

func TestCRDStatusRestoreItemActionExecute(t *testing.T) {
	originalStatus := map[string]interface{}{"phase": "Ready"}
	tests := []struct {
		name          string
		apiVersion    string
		currentStatus map[string]interface{}
		backupStatus  map[string]interface{}
		wantStatus    map[string]interface{}
	}{
		{
			name:          "restores custom resource status from backup",
			apiVersion:    "example.io/v1",
			currentStatus: map[string]interface{}{"phase": "Pending"},
			backupStatus:  originalStatus,
			wantStatus:    originalStatus,
		},
		{
			name:          "does not change built-in resource status",
			apiVersion:    "networking.k8s.io/v1",
			currentStatus: map[string]interface{}{"phase": "Current"},
			backupStatus:  originalStatus,
			wantStatus:    map[string]interface{}{"phase": "Current"},
		},
		{
			name:          "leaves status unchanged when backup has none",
			apiVersion:    "example.io/v1",
			currentStatus: map[string]interface{}{"phase": "Current"},
			wantStatus:    map[string]interface{}{"phase": "Current"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			current := map[string]interface{}{
				"apiVersion": test.apiVersion,
				"kind":       "Example",
				"metadata":   map[string]interface{}{"name": "example"},
			}
			backup := map[string]interface{}{}
			if test.backupStatus != nil {
				backup["status"] = test.backupStatus
			}
			if test.currentStatus != nil {
				current["status"] = test.currentStatus
			}

			output, err := (&CRDStatusRestoreItemAction{Log: logrus.New()}).Execute(executeInput(current, backup))
			if err != nil {
				t.Fatalf("Execute() error = %v", err)
			}
			gotStatus, found, err := unstructured.NestedMap(output.UpdatedItem.UnstructuredContent(), "status")
			if err != nil {
				t.Fatalf("reading updated status: %v", err)
			}
			if test.wantStatus == nil {
				if found {
					t.Fatalf("updated item unexpectedly has status %v", gotStatus)
				}
				return
			}
			if !found || !reflect.DeepEqual(gotStatus, test.wantStatus) {
				t.Errorf("updated status = %v (found %t), want %v", gotStatus, found, test.wantStatus)
			}
		})
	}
}

func TestCRDStatusRestoreItemActionExecuteInvalidBackupStatus(t *testing.T) {
	input := executeInput(
		map[string]interface{}{"apiVersion": "example.io/v1", "metadata": map[string]interface{}{"name": "example"}},
		map[string]interface{}{"status": "invalid"},
	)

	output, err := (&CRDStatusRestoreItemAction{Log: logrus.New()}).Execute(input)
	if err == nil {
		t.Fatal("Execute() error = nil, want error for a non-map backup status")
	}
	if output == nil || output.UpdatedItem == nil {
		t.Fatal("Execute() returned no updated item")
	}
}

func TestResourceCreationTimestampExecute(t *testing.T) {
	tests := []struct {
		name               string
		currentAnnotations map[string]interface{}
		backupTimestamp    string
		wantAnnotations    map[string]string
	}{
		{
			name: "adds original timestamp and preserves annotations",
			currentAnnotations: map[string]interface{}{
				"example.io/keep": "yes",
			},
			backupTimestamp: "2020-01-02T03:04:05Z",
			wantAnnotations: map[string]string{
				"example.io/keep":    "yes",
				AnnPreviousTimestamp: "2020-01-02T03:04:05Z",
			},
		},
		{
			name: "does not overwrite existing timestamp annotation",
			currentAnnotations: map[string]interface{}{
				AnnPreviousTimestamp: "existing",
				"example.io/keep":    "yes",
			},
			backupTimestamp: "2020-01-02T03:04:05Z",
			wantAnnotations: map[string]string{
				AnnPreviousTimestamp: "existing",
				"example.io/keep":    "yes",
			},
		},
		{
			name:            "initializes annotations when backup has no timestamp",
			wantAnnotations: map[string]string{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			metadata := map[string]interface{}{"name": "example"}
			if test.currentAnnotations != nil {
				metadata["annotations"] = test.currentAnnotations
			}
			backup := map[string]interface{}{"metadata": map[string]interface{}{"name": "example"}}
			if test.backupTimestamp != "" {
				backup["metadata"].(map[string]interface{})["creationTimestamp"] = test.backupTimestamp
			}

			output, err := (&ResourceCreationTimestamp{Log: logrus.New()}).Execute(executeInput(
				map[string]interface{}{"metadata": metadata},
				backup,
			))
			if err != nil {
				t.Fatalf("Execute() error = %v", err)
			}
			gotAnnotations, found, err := unstructured.NestedStringMap(output.UpdatedItem.UnstructuredContent(), "metadata", "annotations")
			if err != nil {
				t.Fatalf("reading updated annotations: %v", err)
			}
			if !found || !reflect.DeepEqual(gotAnnotations, test.wantAnnotations) {
				t.Errorf("updated annotations = %v (found %t), want %v", gotAnnotations, found, test.wantAnnotations)
			}
		})
	}
}

func TestResourceCreationTimestampExecuteInvalidAnnotations(t *testing.T) {
	input := executeInput(
		map[string]interface{}{"metadata": map[string]interface{}{"annotations": "invalid"}},
		map[string]interface{}{},
	)

	output, err := (&ResourceCreationTimestamp{Log: logrus.New()}).Execute(input)
	if err == nil {
		t.Fatal("Execute() error = nil, want error for invalid annotations")
	}
	if output == nil || output.UpdatedItem == nil {
		t.Fatal("Execute() returned no updated item")
	}
}

func TestResourceLastKnownAnnotationExecute(t *testing.T) {
	tests := []struct {
		name            string
		annotations     map[string]interface{}
		wantAnnotations map[string]string
	}{
		{
			name: "removes last-applied annotation and preserves others",
			annotations: map[string]interface{}{
				AnnToRemove:       `{"kind":"Example"}`,
				"example.io/keep": "yes",
			},
			wantAnnotations: map[string]string{"example.io/keep": "yes"},
		},
		{
			name:            "leaves resource without annotations unchanged",
			wantAnnotations: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			metadata := map[string]interface{}{"name": "example"}
			if test.annotations != nil {
				metadata["annotations"] = test.annotations
			}
			output, err := (&ResourceLastKnownAnnotation{Log: logrus.New()}).Execute(executeInput(
				map[string]interface{}{"metadata": metadata},
				map[string]interface{}{},
			))
			if err != nil {
				t.Fatalf("Execute() error = %v", err)
			}
			gotAnnotations, found, err := unstructured.NestedStringMap(output.UpdatedItem.UnstructuredContent(), "metadata", "annotations")
			if err != nil {
				t.Fatalf("reading updated annotations: %v", err)
			}
			if test.wantAnnotations == nil {
				if found {
					t.Fatalf("updated item unexpectedly has annotations %v", gotAnnotations)
				}
				return
			}
			if !found || !reflect.DeepEqual(gotAnnotations, test.wantAnnotations) {
				t.Errorf("updated annotations = %v (found %t), want %v", gotAnnotations, found, test.wantAnnotations)
			}
		})
	}
}

func TestResourceLastKnownAnnotationExecuteInvalidAnnotations(t *testing.T) {
	input := executeInput(
		map[string]interface{}{"metadata": map[string]interface{}{"annotations": "invalid"}},
		map[string]interface{}{},
	)

	output, err := (&ResourceLastKnownAnnotation{Log: logrus.New()}).Execute(input)
	if err == nil {
		t.Fatal("Execute() error = nil, want error for invalid annotations")
	}
	if output == nil || output.UpdatedItem == nil {
		t.Fatal("Execute() returned no updated item")
	}
}

func executeInput(current, backup map[string]interface{}) *velero.RestoreItemActionExecuteInput {
	return &velero.RestoreItemActionExecuteInput{
		Item:           &unstructured.Unstructured{Object: current},
		ItemFromBackup: &unstructured.Unstructured{Object: backup},
	}
}
