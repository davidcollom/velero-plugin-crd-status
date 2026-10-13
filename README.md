# Velero CRD Status Plugin

This project provides restore item action plugins for [Velero](https://velero.io/). It helps preserve selected custom-resource information during restore and removes a Kubernetes annotation that can interfere with applying restored resources.

## Restore actions

The plugin registers these Velero restore item actions:

| Plugin name | Behavior |
| --- | --- |
| `jetstack.io/crd-status-restorer` | Copies the `status` field from the backed-up resource onto a restored custom resource. Custom resources are identified heuristically: API versions with a group are treated as custom unless the group is `v1` or ends in `.k8s.io`. |
| `jetstack.io/creation-timestamp-annotation` | Copies the backed-up `metadata.creationTimestamp` into the `jetstack.io/original-timestamp` annotation, without replacing an annotation already present. |
| `jetstack.io/last-known-annotation` | Removes `kubectl.kubernetes.io/last-applied-configuration` from restored resources when present. |

Each action applies to all resources; the status action checks whether a resource is a custom resource before restoring its status.

## Build and install

Build the plugin binary from the repository root:

```sh
go build -o velero-plugin-crd-status ./
```

The included `Dockerfile` packages the built binary for Velero's plugin directory. Build and publish the image, then include it in the Velero installation's `--plugins` list alongside the provider plugin. For example:

```sh
docker build -t <registry>/velero-plugin-crd-status:<tag> .
velero install --plugins <provider-plugin-image>,<registry>/velero-plugin-crd-status:<tag>
```

Use the action names in the table above when configuring Velero restore item actions for your environment.

## Examples

`examples/examples.yaml` contains a sample Velero schedule, namespace, and custom resource. It illustrates a resource with a `status` field that the status restore action can preserve.

## Development

Run the unit tests with:

```sh
go test ./...
```

The `Makefile` also contains Terraform, kind, and Velero commands for the project's example integration environment. Those targets require the corresponding tools and cloud credentials.
