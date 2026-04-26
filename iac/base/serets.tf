# version_destroy_ttl keeps superseded versions in DESTROYED state after the TTL,
# so only 1 active version exists per secret at steady state — staying free tier.
locals {
  secret_version_destroy_ttl = "86400s" # 1 day
}

resource "google_secret_manager_secret" "api_fetcher_psk" {
  secret_id           = "api-fetcher-psk"
  version_destroy_ttl = local.secret_version_destroy_ttl

  labels = {
    label = "4ks"
  }

  replication {
    user_managed {
      replicas {
        location = var.region
      }
    }
  }
}

resource "google_secret_manager_secret" "auth0_client_secret" {
  secret_id           = "auth0-client-secret"
  version_destroy_ttl = local.secret_version_destroy_ttl

  labels = {
    label = "4ks"
  }

  replication {
    user_managed {
      replicas {
        location = var.region
      }
    }
  }
}

resource "google_secret_manager_secret" "auth0_session_secret" {
  secret_id           = "auth0-session-secret"
  version_destroy_ttl = local.secret_version_destroy_ttl

  labels = {
    label = "4ks"
  }

  replication {
    user_managed {
      replicas {
        location = var.region
      }
    }
  }
}

resource "google_secret_manager_secret" "typesense_api_key" {
  secret_id           = "typesense-api-key"
  version_destroy_ttl = local.secret_version_destroy_ttl

  labels = {
    label = "4ks"
  }

  replication {
    user_managed {
      replicas {
        location = var.region
      }
    }
  }
}

resource "google_secret_manager_secret" "typesense_search_api_key" {
  secret_id           = "typesense-search-api-key"
  version_destroy_ttl = local.secret_version_destroy_ttl

  labels = {
    label = "4ks"
  }

  replication {
    user_managed {
      replicas {
        location = var.region
      }
    }
  }
}

