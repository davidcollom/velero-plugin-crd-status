
variable "cluster_name" {
  type    = "string"
  default = "velero"
}

output "serviceaccount_key" {
  value     = base64decode(google_service_account_key.velero_kind.private_key)
  sensitive = true
}
