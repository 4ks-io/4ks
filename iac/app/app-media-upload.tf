
data "archive_file" "media_upload" {
  type        = "zip"
  source_dir  = "${path.module}/../../apps/media-upload"
  output_path = "/tmp/media-upload.zip"
}

resource "google_storage_bucket_object" "media_upload_zip" {
  source       = data.archive_file.media_upload.output_path
  content_type = "application/zip"
  name         = "src-${data.archive_file.media_upload.output_md5}.zip"
  bucket       = data.google_storage_bucket.media_upload.name
}

resource "google_cloudfunctions2_function" "media_upload" {
  name        = "media-upload"
  location    = var.region
  description = "creates image upload post validation"

  build_config {
    runtime     = "go126"
    entry_point = "UploadImage"
    source {
      storage_source {
        bucket = data.google_storage_bucket.media_upload.name
        object = google_storage_bucket_object.media_upload_zip.name
      }
    }
  }

  // https://cloud.google.com/functions/pricing
  service_config {
    # min_instance_count = 1
    max_instance_count             = 100
    available_memory               = "1024Mi"
    timeout_seconds                = 15
    ingress_settings               = "ALLOW_INTERNAL_ONLY"
    all_traffic_on_latest_revision = true
    service_account_email          = data.google_service_account.media_upload.email
    environment_variables = {
      DISTRIBUTION_BUCKET  = data.google_storage_bucket.media_read.name
      FIRESTORE_PROJECT_ID = "${local.stage}-${local.org}"
    }
  }

  event_trigger {
    trigger_region = "us"
    event_type     = "google.cloud.storage.object.v1.finalized"
    # https://cloud.google.com/functions/docs/bestpractices/retries
    # retry_policy          = "RETRY_POLICY_RETRY"
    retry_policy          = "RETRY_POLICY_DO_NOT_RETRY"
    service_account_email = data.google_service_account.media_upload.email
    event_filters {
      attribute = "bucket"
      value     = data.google_storage_bucket.media_write.name
    }
  }
}
