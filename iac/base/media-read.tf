resource "google_storage_bucket" "media_read" {
  name          = "media-read.${local.web_domain}"
  location      = "us"
  force_destroy = true

  uniform_bucket_level_access = true

  cors {
    origin = compact([
      terraform.workspace == "base-dev-us-east" ? "https://local.4ks.io" : "",
      "https://${local.www_domain}",
    ])
    method          = ["GET"]
    response_header = ["*"]
    max_age_seconds = 3600
  }
}


resource "google_storage_bucket_iam_member" "media_read_viewer" {
  bucket = google_storage_bucket.media_read.name
  role   = "roles/storage.objectViewer"
  member = "allUsers"
}
