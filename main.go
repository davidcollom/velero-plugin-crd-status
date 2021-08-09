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

package main

import (
	"github.com/davidcollom/velero-plugin-crd-status/internal/restore"
	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	veleroplugin "github.com/vmware-tanzu/velero/pkg/plugin/framework"
)

func main() {
	veleroplugin.NewServer().
		BindFlags(pflag.CommandLine).
		RegisterRestoreItemAction("jetstack.io/crd-status-restorer", newCRDStatusRestoreItemAction).
		RegisterRestoreItemAction("jetstack.io/creation-timestamp-annotation", newResourceCreationTimestamp).
		RegisterRestoreItemAction("jetstack.io/last-known-annotation", newResourceLastKnownAnnotation).
		Serve()
}

func newCRDStatusRestoreItemAction(logger logrus.FieldLogger) (interface{}, error) {
	return &restore.CRDStatusRestoreItemAction{Log: logger}, nil
}

func newResourceCreationTimestamp(logger logrus.FieldLogger) (interface{}, error) {
	return &restore.ResourceCreationTimestamp{Log: logger}, nil
}
func newResourceLastKnownAnnotation(logger logrus.FieldLogger) (interface{}, error) {
	return &restore.ResourceLastKnownAnnotation{Log: logger}, nil
}
