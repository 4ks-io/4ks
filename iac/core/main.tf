terraform {
  required_version = ">= 1.9.0"
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 6.0"
    }
  }
  
  cloud {
    organization = "4ks"
    workspaces {
      # https://www.terraform.io/cli/cloud/settings
      tags = ["core","us-east"]
    }
  }
}


provider "google" {
  project = local.project
  region  = var.region
  zone    = var.zone
}

