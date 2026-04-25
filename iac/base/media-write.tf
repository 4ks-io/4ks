resource "google_storage_bucket" "media_write" {
  name          = "media-write.${local.web_domain}"
  location      = "us"
  force_destroy = local.stage != "prd"

  // SBX upload validation passed with uniform bucket-level access enabled, so
  // promote the same bucket-level IAM model into the shared environments.
  uniform_bucket_level_access = true

  cors {
    origin = compact([
      terraform.workspace == "base-dev-us-east" ? "https://local.4ks.io" : "",
      "https://${local.www_domain}",
    ])
    method          = ["PUT"]
    response_header = ["*"]
    max_age_seconds = 3600
  }

  lifecycle_rule {
    condition {
      age = 1
    }
    action {
      type = "Delete"
    }
  }
}
