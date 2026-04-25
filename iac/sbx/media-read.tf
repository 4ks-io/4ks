resource "google_storage_bucket" "media_read" {
  name          = "media-read.${local.web_domain}"
  location      = var.region
  force_destroy = true

  uniform_bucket_level_access = true

  cors {
    origin          = ["https://local.4ks.io"]
    method          = ["GET"]
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


resource "google_storage_bucket_iam_member" "media_read_viewer" {
  bucket = google_storage_bucket.media_read.name
  role   = "roles/storage.objectViewer"
  # trivy:ignore:AVD-GCP-0001 Sandbox media read bucket is intentionally public.
  member = "allUsers"
}
