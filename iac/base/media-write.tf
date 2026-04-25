resource "google_storage_bucket" "media_write" {
  name          = "media-write.${local.web_domain}"
  location      = "us"
  force_destroy = local.stage != "prd"

  uniform_bucket_level_access = false

  cors {
    // should be google_cloudfunctions2_function.media_upload.service_config[0].uri
    # origin          = ["https:/${local.web_domain}", "https://media-upload-*.a.run.app"]
    origin          = ["*"]
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
