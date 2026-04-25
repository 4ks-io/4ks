# iam
resource "google_service_account" "api" {
  account_id   = "cloud-run-api-sa"
  display_name = "Cloud Run API"
  description  = "Identity used by the Cloud Run API service"
}

resource "google_project_iam_member" "service_agent" {
  project = local.project
  role    = "roles/appengine.serviceAgent"
  member  = "serviceAccount:${google_service_account.api.email}"
}

resource "google_project_iam_member" "datastore_user" {
  project = local.project
  role    = "roles/datastore.user"
  member  = "serviceAccount:${google_service_account.api.email}"
}

resource "google_project_iam_member" "firestore_service_agent" {
  project = local.project
  role    = "roles/firestore.serviceAgent"
  member  = "serviceAccount:${google_service_account.api.email}"
}

resource "google_project_iam_member" "pubsub_service_agent" {
  project = local.project
  role    = "roles/pubsub.editor"
  member  = "serviceAccount:${google_service_account.api.email}"
}

resource "google_project_iam_custom_role" "api" {
  role_id     = "cloudRunAPI"
  title       = "Cloud Run API"
  description = "Cloud Run API"
  permissions = [
    "storage.buckets.get",
    "storage.objects.create",
    "storage.objects.delete"
  ]
}

data "google_iam_policy" "api_custom" {
  binding {
    role = google_project_iam_custom_role.api.id
    members = [
      "serviceAccount:${google_service_account.api.email}",
    ]
  }
}

data "google_iam_policy" "api_token_creator" {
  binding {
    role = "roles/iam.serviceAccountTokenCreator"
    members = [
      "serviceAccount:${google_service_account.api.email}",
    ]
  }
}

// Add the narrower upload-bucket grant before removing the older project-wide
// custom storage role in a later phase.
resource "google_storage_bucket_iam_member" "api_media_write_creator" {
  bucket = data.google_storage_bucket.media_write.name
  role   = "roles/storage.objectCreator"
  member = "serviceAccount:${google_service_account.api.email}"
}

// The API reads fallback media metadata from the static bucket, so add a
// bucket-scoped viewer grant before tightening broader storage access.
resource "google_storage_bucket_iam_member" "api_media_static_viewer" {
  bucket = data.google_storage_bucket.media_static.name
  role   = "roles/storage.objectViewer"
  member = "serviceAccount:${google_service_account.api.email}"
}

// Add the topic-scoped publisher grant first; keep the broader Pub/Sub role in
// place until fetch publishing has been revalidated in runtime environments.
resource "google_pubsub_topic_iam_member" "api_fetcher_publisher" {
  project = local.project
  topic   = google_pubsub_topic.fetcher.name
  role    = "roles/pubsub.publisher"
  member  = "serviceAccount:${google_service_account.api.email}"
}

# service
resource "google_cloud_run_v2_service" "api" {
  name     = "api"
  location = var.region
  traffic {
    percent = 100
    type    = "TRAFFIC_TARGET_ALLOCATION_TYPE_LATEST"
  }

  template {
    scaling {
      min_instance_count = 0
      max_instance_count = 10
    }

    service_account = google_service_account.api.email
    containers {
      image = "${local.container_registry}/api/app:${var.api_build_number}"

      resources {
        cpu_idle = true
        limits = {
          cpu    = "1000m"
          memory = "128Mi"
        }
      }

      ports {
        container_port = 5000
      }

      env {
        name  = "SERVICE_ACCOUNT_EMAIL"
        value = google_service_account.api.email
      }

      env {
        name  = "UPLOADABLE_BUCKET"
        value = data.google_storage_bucket.media_write.name
      }
      env {
        name  = "DISTRIBUTION_BUCKET"
        value = data.google_storage_bucket.media_read.name
      }
      env {
        name  = "STATIC_MEDIA_BUCKET"
        value = data.google_storage_bucket.media_static.name
      }
      env {
        name  = "GOOGLE_CLOUD_PROJECT"
        value = data.google_project.project.number
      }

      env {
        name  = "FIRESTORE_PROJECT_ID"
        value = "${local.stage}-${local.org}"
      }
      env {
        name  = "AUTH0_DOMAIN"
        value = var.auth0_domain[terraform.workspace]
      }
      env {
        name  = "AUTH0_AUDIENCE"
        value = local.api_url
      }
      env {
        name = "CORS_ALLOWED_ORIGINS"
        value = join(",", compact([
          terraform.workspace == "app-dev-us-east" ? "https://local.4ks.io" : "",
          local.web_url,
        ]))
      }
      env {
        name  = "TRUSTED_PROXY_CIDRS"
        value = "10.0.0.0/8,172.16.0.0/12,192.168.0.0/16"
      }
      env {
        name  = "EXPORTER_TYPE"
        value = "JAEGER"
      }
      env {
        name  = "GIN_MODE"
        value = "release"
      }
      env {
        name  = "JAEGER_ENABLED"
        value = var.feat_api_jaeger_enabled_map[terraform.workspace]
      }
      env {
        name  = "SWAGGER_ENABLED"
        value = var.feat_api_swagger_enabled_map[terraform.workspace]
      }
      env {
        name  = "SWAGGER_URL_PREFIX"
        value = "/api"
      }

      env {
        name  = "TYPESENSE_URL"
        value = "https://${var.typesense_url_env_map[terraform.workspace]}"
      }

      env {
        name  = "TYPESENSE_API_KEY"
        value = var.typesense_api_key
      }
      env {
        name  = "MEDIA_FALLBACK_URL"
        value = "${local.web_url}"
      }
      env {
        name  = "STATIC_MEDIA_FALLBACK_PREFIX"
        value = "static/fallback/f"
      }
      env {
        name  = "MEDIA_IMAGE_URL"
        value = "${local.web_url}/image"
      }

      env {
        name  = "PUBSUB_PROJECT_ID"
        value = "${local.stage}-${local.org}"
      }
      env {
        name  = "API_FETCHER_PSK"
        value = data.google_secret_manager_secret_version.api_fetcher_psk.secret_data

      }
    }
  }
}



resource "google_cloud_run_service_iam_member" "api_anonymous_access" {
  service  = google_cloud_run_v2_service.api.name
  location = google_cloud_run_v2_service.api.location
  role     = "roles/run.invoker"
  member   = "allUsers"
}

resource "google_compute_region_network_endpoint_group" "api_neg" {
  name                  = "${local.project}-api-neg"
  network_endpoint_type = "SERVERLESS"
  region                = var.region
  cloud_run {
    service = google_cloud_run_v2_service.api.name
  }
}

# resource "google_compute_backend_service" "api" {
#   name = "${local.project}-api-backend"

#   protocol    = "HTTP"
#   port_name   = "http"
#   timeout_sec = 30

#   backend {
#     group = google_compute_region_network_endpoint_group.api_neg.id
#   }
# }

output "api_service_url" {
  value = google_cloud_run_v2_service.api.uri
}
