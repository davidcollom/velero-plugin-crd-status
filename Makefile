SECRET_FILE ?= service-account.json
KIND_CLUSTER_NAME ?= velero

.PHONY: terraform-init
terraform-init: terraform-fmt
	cd terraform; terraform init .

.PHONY: terraform-apply
terraform-apply: terraform-init
	cd terraform; terraform apply --auto-approve --var=cluster_name=$(KIND_CLUSTER_NAME) .

.PHONY: terraform-plan
terraform-plan: terraform-init
	cd terraform; terraform plan --var=cluster_name=$(KIND_CLUSTER_NAME) .

.PHONY: terraform-fmt
terraform-fmt:
	terraform fmt terraform/

.PHONY: velero-install
velero-install: serviceaccount.json
	velero install --provider gcp --bucket david-jetstack-velero-testing --plugins velero/velero-plugin-for-gcp:v1.0.0,davidcollom/velero-plugin-crd-status:latest --wait --secret-file ./serviceaccount.json

serviceaccount.json: terraform-apply
	cd terraform; terraform output serviceaccount_key > ../serviceaccount.json

.PHONY: context
context:
	kubectx kind-$(KIND_CLUSTER_NAME)

.PHONY: setup
setup: terraform-apply build velero-install deploy

.PHONY: build
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
	kind --name=$(KIND_CLUSTER_NAME) load docker-image davidcollom/velero-plugin-crd-status:latest

.PHONY: restart
restart: build
	kubectl delete pods --all -n velero

.PHONY: deploy-crds
deploy-crds: context
	kubectl apply -k github.com/argoproj/argo-events/manifests/base/crds/?ref=master
	kubectl apply -k github.com/argoproj/argo-workflows/manifests/base/crds/minimal?ref=master
	kubectl apply -f examples/crds.yaml

.PHONY: deploy
deploy: deploy-crds
	kubectl apply -f examples/examples.yaml

.PHONY: backup
backup: context
	velero backup create --from-schedule=system

.PHONY: restore
restore: context
	velero restore create --from-schedule=system

.PHONY: modules
modules:
	go mod tidy
