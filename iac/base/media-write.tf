resource "google_storage_bucket" "media_write" {
  name          = "media-write.${local.web_domain}"
  location      = "us"
  force_destroy = local.stage != "prd"

  uniform_bucket_level_access = false

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
