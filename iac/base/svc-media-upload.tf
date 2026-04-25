resource "google_storage_bucket" "media_upload" {
  name     = "${local.org}-${local.stage}-media-upload-deploy"
  location = var.region
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

resource "google_project_iam_member" "datastore_reader" {
  project    = local.project
  role       = "roles/datastore.user"
  member     = "serviceAccount:${google_service_account.media_upload.email}"
  depends_on = [google_project_iam_member.artifactregistry_reader]
}

// Add the narrower bucket-scoped source permission before removing the older
// project-wide custom role binding in a later phase.
resource "google_storage_bucket_iam_member" "media_upload_source_admin" {
  bucket     = google_storage_bucket.media_write.name
  role       = "roles/storage.objectAdmin"
  member     = "serviceAccount:${google_service_account.media_upload.email}"
  depends_on = [google_project_iam_member.datastore_reader]
}

// The upload function also needs object-level access on the processed-media
// bucket so it can write resized variants during the staged IAM migration.
// With the bucket-scoped grants in place, the function should depend on the
// narrower storage bindings rather than the old project-level custom role.
resource "google_storage_bucket_iam_member" "media_upload_distribution_admin" {
  bucket     = google_storage_bucket.media_read.name
  role       = "roles/storage.objectAdmin"
  member     = "serviceAccount:${google_service_account.media_upload.email}"
  depends_on = [google_project_iam_member.datastore_reader]
}
