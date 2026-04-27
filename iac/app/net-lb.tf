# resource "google_compute_project_default_network_tier" "default" {
#   network_tier = "PREMIUM"
# }

# module "lb" {
#   source                = "./modules/http-global-load-balancer"
#   name                  = local.project
#   project               = data.google_project.project.number
#   # region                = var.region
#   url_map               = google_compute_url_map.urlmap.self_link
#   dns_managed_zone_name = local.dns_managed_zone_name
#   custom_domain_names   = [local.web_domain]
#   create_dns_entries    = var.create_dns_entry
#   dns_record_ttl        = var.dns_record_ttl
#   enable_http           = true
#   enable_ssl            = true

#   # hack to prevent the resource from constantly changing
#   # ssl_certificates      = [google_compute_managed_ssl_certificate.default.id]
#   ssl_certificates = ["https://www.googleapis.com/compute/v1/projects/${local.stage}-${local.org}/global/sslCertificates/${local.google_compute_managed_ssl_certificate_name}"]

#   #   custom_labels = var.custom_labels
# }


# resource "google_compute_url_map" "urlmap" {
#   project = data.google_project.project.number

#   name        = "${local.project}-url-map"
#   description = "URL map for ${local.project}"

#   default_service = google_compute_backend_service.web.id

#   host_rule {
#     hosts        = ["*"]
#     path_matcher = "all"
#   }

#   path_matcher {
#     name            = "all"
#     default_service = google_compute_backend_service.web.id

#     path_rule {
#       paths   = ["/api/*"]
#       service = google_compute_backend_service.api.id
#       route_action {
#         url_rewrite {
#           path_prefix_rewrite = "/"
#         }
#       }
#     }

#     path_rule {
#       paths   = ["/static/*"]
#       service = data.google_compute_backend_bucket.media_static.id
#       route_action {
#         url_rewrite {
#           path_prefix_rewrite = "/"
#         }
#       }
#     }
#   }

#   test {
#     service = google_compute_backend_service.api.id
#     host    = local.web_domain
#     path    = "/api/ready"
#   }

# }

# resource "google_compute_security_policy" "development" {
#   name = "${local.stage}-policy"

#   # rule {
#   #   action   = "allow"
#   #   priority = "1000"
#   #   match {
#   #     versioned_expr = "SRC_IPS_V1"
#   #     config {
#   #       src_ip_ranges = var.development_ip_addresses
#   #     }
#   #   }
#   #   description = "Allow access to non-production sites"
#   # }

#   rule {
#     # action = "deny(403)"
#     action   = "allow"
#     priority = "2147483647"
#     match {
#       versioned_expr = "SRC_IPS_V1"
#       config {
#         src_ip_ranges = ["*"]
#       }
#     }
#     description = "default rule"
#   }
# }
