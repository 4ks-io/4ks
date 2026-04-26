resource "google_service_account" "web" {
  account_id   = "cloud-run-web-sa"
  display_name = "Cloud Run Web"
  description  = "Identity used by the Cloud Run web service"
}

resource "google_cloud_run_v2_service" "web" {
  name     = "web"
  location = var.region
  # Public web service — accessible by end users via load balancer and run.app URL.
  ingress = "INGRESS_TRAFFIC_ALL"
  traffic {
    percent = 100
    type    = "TRAFFIC_TARGET_ALLOCATION_TYPE_LATEST"
  }

  template {
    scaling {
      min_instance_count = 0
      max_instance_count = 2
    }

    service_account = google_service_account.web.email

    containers {
      image = "${local.container_registry}/web/app:${var.web_build_number}"

      resources {
        cpu_idle = true
        limits = {
          cpu    = "1000m"
          memory = "512Mi"
        }
      }

      ports {
        container_port = 5000
      }
      env {
        name  = "APP_ENV"
        value = var.app_env_map[terraform.workspace]
      }
      env {
        name  = "IO_4KS_API_URL"
        value = var.api_cloudrun_url_env_map[terraform.workspace]
      }
      env {
        name  = "LOG_DEBUG"
        value = "false"
      }
      # auth0
      env {
        name  = "APP_BASE_URL"
        value = local.web_url
      }
      env {
        name  = "AUTH0_CLIENT_ID"
        value = var.auth0_client_id_env_map[terraform.workspace]
      }
      env {
        name = "AUTH0_CLIENT_SECRET"
        value_source {
          secret_key_ref {
            secret  = "auth0-client-secret"
            version = "latest"
          }
        }
      }
      env {
        name = "AUTH0_SECRET"
        value_source {
          secret_key_ref {
            secret  = "auth0-session-secret"
            version = "latest"
          }
        }
      }
      env {
        name  = "AUTH0_DOMAIN"
        value = var.auth0_domain[terraform.workspace]
      }
      env {
        name  = "AUTH0_AUDIENCE"
        value = local.api_url
      }
      # typesense
      env {
        name  = "TYPESENSE_URL"
        value = var.typesense_url_env_map[terraform.workspace]
      }
      env {
        name  = "TYPESENSE_PROTOCOL"
        value = "https"
      }
      env {
        name  = "TYPESENSE_PORT"
        value = "443"
      }
      env {
        name  = "TYPESENSE_URL_CLIENT"
        value = var.typesense_url_env_map[terraform.workspace]
      }
      env {
        name = "TYPESENSE_API_KEY"
        value_source {
          secret_key_ref {
            secret  = "typesense-search-api-key"
            version = "latest"
          }
        }
      }
      env {
        name  = "MEDIA_FALLBACK_URL"
        value = "${local.web_url}/static/fallback"
      }

    }
  }

}

resource "google_cloud_run_v2_service_iam_member" "web_anonymous_access" {
  location = google_cloud_run_v2_service.web.location
  name     = google_cloud_run_v2_service.web.name
  role     = "roles/run.invoker"
  member   = "allUsers"
}

resource "google_secret_manager_secret_iam_member" "web_auth0_client_secret" {
  secret_id = "auth0-client-secret"
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:${google_service_account.web.email}"
}

resource "google_secret_manager_secret_iam_member" "web_auth0_session_secret" {
  secret_id = "auth0-session-secret"
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:${google_service_account.web.email}"
}

resource "google_secret_manager_secret_iam_member" "web_typesense_search_api_key" {
  secret_id = "typesense-search-api-key"
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:${google_service_account.web.email}"
}


resource "google_compute_region_network_endpoint_group" "web_neg" {
  name                  = "${local.project}-web-neg"
  network_endpoint_type = "SERVERLESS"
  region                = var.region
  cloud_run {
    service = google_cloud_run_v2_service.web.name
  }
}

# resource "google_compute_backend_service" "web" {
#   name        = "${local.project}-web-backend"
#   protocol    = "HTTP"
#   port_name   = "http"
#   timeout_sec = 30

#   security_policy = google_compute_security_policy.development.id

#   backend {
#     group = google_compute_region_network_endpoint_group.web_neg.id
#   }
# }

output "web_service_url" {
  value = google_cloud_run_v2_service.web.uri
}
