
terraform {
  required_providers {
    kind = {
      source  = "kyma-incubator/kind"
      version = "0.0.9"
    }
  }
}

provider "google" {
  version = "~> 3.78.0"
  project = "jetstack-david"
  region  = "europe-west2"
}

provider "kind" {}
