resource "google_storage_bucket" "media_upload" {
  name     = "${local.org}-${var.stage}-media-upload-deploy"
  location = var.region
}

data "archive_file" "media_upload" {
  type        = "zip"
  source_dir  = "${path.module}/../../apps/media-upload"
  output_path = "/tmp/media-upload.zip"
}

resource "google_storage_bucket_object" "media_upload_zip" {
  source       = data.archive_file.media_upload.output_path
  content_type = "application/zip"
  name         = "src-${data.archive_file.media_upload.output_md5}.zip"
  bucket       = google_storage_bucket.media_upload.name
}

# To use GCS CloudEvent triggers, the GCS service account requires the Pub/Sub Publisher(roles/pubsub.publisher) IAM role in the specified project.
# (See https://cloud.google.com/eventarc/docs/run/quickstart-storage#before-you-begin)
resource "google_project_iam_member" "gcs_pubsub_publishing" {
  project = local.project
  role    = "roles/pubsub.publisher"
  member  = "serviceAccount:${data.google_storage_project_service_account.gcs_account.email_address}"
}

resource "google_service_account" "media_upload" {
  account_id   = "media-upload-sa"
  display_name = "Service Account to Create Media upload"
}

# Permissions on the service account used by the function and Eventarc trigger
resource "google_project_iam_member" "invoking" {
  project    = local.project
  role       = "roles/run.invoker"
  member     = "serviceAccount:${google_service_account.media_upload.email}"
  depends_on = [google_project_iam_member.gcs_pubsub_publishing]
}

resource "google_project_iam_member" "event_receiving" {
  project    = local.project
  role       = "roles/eventarc.eventReceiver"
  member     = "serviceAccount:${google_service_account.media_upload.email}"
  depends_on = [google_project_iam_member.invoking]
}

resource "google_project_iam_member" "artifactregistry_reader" {
  project    = local.project
  role       = "roles/artifactregistry.reader"
  member     = "serviceAccount:${google_service_account.media_upload.email}"
  depends_on = [google_project_iam_member.event_receiving]
}

resource "google_project_iam_member" "bucket" {
  project    = local.project
  role       = google_project_iam_custom_role.bucket.name
  member     = "serviceAccount:${google_service_account.media_upload.email}"
  depends_on = [google_project_iam_member.artifactregistry_reader]
}

resource "google_project_iam_member" "datastore_reader" {
  project    = local.project
  role       = "roles/datastore.user"
  member     = "serviceAccount:${google_service_account.media_upload.email}"
  depends_on = [google_project_iam_member.artifactregistry_reader]
}

resource "google_cloudfunctions2_function" "media_upload" {
  depends_on = [
    google_project_iam_member.bucket,
    google_project_iam_member.event_receiving,
    google_project_iam_member.artifactregistry_reader,
  ]

  name        = "media-upload"
  location    = var.region
  description = "creates image upload post validation"

  build_config {
    runtime     = "go126"
    entry_point = "UploadImage"
    source {
      storage_source {
        bucket = google_storage_bucket.media_upload.name
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
    service_account_email          = google_service_account.media_upload.email
    environment_variables = {
      DISTRIBUTION_BUCKET  = google_storage_bucket.media_read.name
      FIRESTORE_PROJECT_ID = "${var.stage}-${local.org}"
      IO_4KS_DEVELOPMENT = "true"
    }
  }

  event_trigger {
    trigger_region = var.region
    event_type     = "google.cloud.storage.object.v1.finalized"
    # https://cloud.google.com/functions/docs/bestpractices/retries
    # retry_policy          = "RETRY_POLICY_RETRY"
    retry_policy          = "RETRY_POLICY_DO_NOT_RETRY"
    service_account_email = google_service_account.media_upload.email
    event_filters {
      attribute = "bucket"
      value     = google_storage_bucket.media_write.name
    }
  }
}
# [END functions_v2_full]