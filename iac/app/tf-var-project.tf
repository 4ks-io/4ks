
// us-east1 : Moncks Corner, South Carolina
variable "region" {
  default = "us-east4"
}

variable "zone" {
  default = "us-east4-b"
}

variable "fetcher_build_number" {
  type = string
}

variable "api_build_number" {
  type = string
}

variable "web_build_number" {
  type = string
}

variable "typesense_api_key" {
  type        = string
  description = "Typesense admin API key. Set as a sensitive Terraform Cloud or CI variable."
  sensitive   = true
}

variable "auth0_client_secret" {
  type        = string
  description = "Auth0 application client secret. Set as a sensitive Terraform Cloud or CI variable."
  sensitive   = true
}

variable "auth0_secret" {
  type        = string
  description = "Auth0 session secret. Set as a sensitive Terraform Cloud or CI variable."
  sensitive   = true
}

variable "auth0_domain" {
  type = map(string)
  default = {
    "app-dev-us-east" = "dev-4ks.us.auth0.com",
    "app-tst-us-east" = ""
    "app-prd-us-east" = "4ks.us.auth0.com",
  }
}

locals {
  org                   = "4ks"
  domain                = "4ks.io"
  stage                 = substr(terraform.workspace, 4, 3)
  dns_managed_zone_name = "4ks-${local.stage}-app-zone"
  project               = "${local.stage}-4ks"
  www_domain            = "${var.www_subdomain_env_map[terraform.workspace]}${local.org}.io"
  web_domain            = "${var.web_subdomain_env_map[terraform.workspace]}${local.org}.io"
  web_url               = "https://${local.www_domain}"
  api_url               = "https://${local.www_domain}/api"
  container_registry    = "us-east4-docker.pkg.dev/eng-${local.org}"

  google_compute_managed_ssl_certificate_name = "managed-certificate"
}
