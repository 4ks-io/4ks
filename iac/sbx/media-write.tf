resource "google_storage_bucket" "media_write" {
  name          = "media-write.${local.web_domain}"
  location      = var.region
  force_destroy = true

  uniform_bucket_level_access = false

  cors {
    origin = [
      "https://local.4ks.io",
      "https://${local.web_domain}",
    ]
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
