resource "google_storage_bucket" "velero_backups" {
  name          = "david-jetstack-velero-testing"
  location      = "europe-west2"
  storage_class = "STANDARD"
  force_destroy = true

  lifecycle_rule {
    condition {
      age = 14
    }

    action {
      type = "Delete"
    }
  }

  retention_policy {
    is_locked        = false
    retention_period = 1209600
  }
}

resource "google_service_account" "velero" {
  account_id   = "velero"
  display_name = "Velero Service Account"
}

resource "google_storage_bucket_iam_binding" "binding" {
  bucket = google_storage_bucket.velero_backups.name
  role   = "roles/storage.admin"
  members = [
    "serviceAccount:${google_service_account.velero.email}"
  ]
}

resource "google_service_account_key" "velero_kind" {
  service_account_id = google_service_account.velero.name
}

output "serviceaccount_key" {
  value = base64decode(google_service_account_key.velero_kind.private_key)
}
