SECRET_FILE ?= service-account.json


terraform-init: terraform-fmt
	cd terraform; terraform init .

terraform-apply: terraform-init
	cd terraform; terraform apply .

terraform-plan: terraform-init
	cd terraform; terraform plan .

terraform-fmt:
	terraform fmt terraform/

velero-install: serviceaccount.json
	velero install --provider gcp --bucket david-jetstack-velero-testing --plugins velero/velero-plugin-for-gcp:v1.0.0,davidcollom/velero-plugin-crd-status:latest --wait --secret-file ./serviceaccount.json

serviceaccount.json: terraform-apply
	cd terraform; terraform output serviceaccount_key > ../serviceaccount.json

context:
	kubectx kind-kind

setup: terraform-apply build velero-install deploy-crds

build:
	CGO_ENABLED=0 \
		GOOS=linux \
		GOARCH=amd64 \
		go build \
    -o velero-plugin-crd-status \
    -installsuffix "static" \
    -mod=readonly \
    ./
	docker build -t davidcollom/velero-plugin-crd-status:latest .
	kind load docker-image davidcollom/velero-plugin-crd-status:latest

deploy-crds:
	kubectl apply -k github.com/argoproj/argo-events/manifests/base/crds/?ref=master§

backup:
	velero backup create --from-schedule=system

restore:
	velero restore create --from-schedule=system


.PHONY: modules
modules:
	go mod tidy
