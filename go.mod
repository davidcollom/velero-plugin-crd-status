module github.com/davidcollom/velero-plugin-crd-status

go 1.13

require (
	github.com/hashicorp/go-hclog v0.12.0 // indirect
	github.com/hashicorp/go-plugin v1.0.1-0.20190610192547-a1bc61569a26 // indirect
	github.com/hashicorp/yamux v0.0.0-20181012175058-2f1d1f20f75d // indirect
	github.com/sirupsen/logrus v1.6.0
	github.com/spf13/pflag v1.0.5
	github.com/vmware-tanzu/velero v1.6.2
	k8s.io/apimachinery v0.19.12
)

replace github.com/gogo/protobuf => github.com/gogo/protobuf v1.3.2
