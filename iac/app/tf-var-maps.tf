variable "app_env_map" {
  type = map(string)
  default = {
    "app-dev-us-east" = "dev"
    "app-tst-us-east" = "tst"
    "app-prd-us-east" = "prd"
  }
}

variable "web_subdomain_env_map" {
  type = map(string)
  default = {
    "app-dev-us-east" = "dev.",
    "app-tst-us-east" = "tst."
    "app-prd-us-east" = ""
  }
}
variable "www_subdomain_env_map" {
  type = map(string)
  default = {
    "app-dev-us-east" = "dev.",
    "app-tst-us-east" = "tst."
    "app-prd-us-east" = "www."
  }
}

variable "feat_api_jaeger_enabled_map" {
  type = map(string)
  default = {
    "app-dev-us-east" = false
    "app-prd-us-east" = false
  }
}

variable "feat_api_swagger_enabled_map" {
  type = map(string)
  default = {
    "app-dev-us-east" = true
    "app-prd-us-east" = false
  }
}

# typesense
variable "typesense_client_api_key_env_map" {
  type        = map(string)
  description = "Typesense search-only client API keys by Terraform workspace. Set as a sensitive Terraform Cloud variable."
  sensitive   = true
}

variable "typesense_url_env_map" {
  type = map(string)
  default = {
    "app-dev-us-east" = "m5wzvue301hkiy9gp-1.a1.typesense.net",
    "app-tst-us-east" = ""
    "app-prd-us-east" = "4d02iphk3zbtgo6jp-1.a1.typesense.net"
  }
}

# auth0
variable "auth0_client_id_env_map" {
  type = map(string)
  default = {
    "app-dev-us-east" = "eM5Zyyfp6coLg3zORMFsZEnQmpqxBjHd"
    "app-tst-us-east" = "tst"
    "app-prd-us-east" = "vAUr50Saqug9Mf3Yu4cFvAaT2nsgNLIN"
  }
}

variable "api_cloudrun_url_env_map" {
  type = map(string)
  default = {
    "app-dev-us-east" = "https://foobar",
    "app-tst-us-east" = "https://foobar"
    "app-prd-us-east" = "https://api-gl2loyoupq-uk.a.run.app"
  }
}
