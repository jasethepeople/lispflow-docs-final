terraform {
  required_providers {
    google = { source = "hashicorp/google", version = "~> 5.0" }
    kubernetes = { source = "hashicorp/kubernetes", version = "~> 2.30" }
    helm = { source = "hashicorp/helm", version = "~> 2.12" }
  }
}

resource "google_container_cluster" "primary" {
  name     = "${var.cluster_name}-${var.environment}"
  location = var.region
  initial_node_count = var.node_count
  deletion_protection = var.environment == "prod" ? true : false

  node_config {
    machine_type = var.machine_type
    oauth_scopes = ["https://www.googleapis.com/auth/cloud-platform"]
  }

  workload_identity_config { workload_pool = "${var.project_id}.svc.id.goog" }
}

resource "google_sql_database_instance" "postgres" {
  name             = "lispflow-${var.environment}"
  database_version = "POSTGRES_15"
  region           = var.region
  deletion_protection = var.environment == "prod"

  settings {
    tier = var.db_tier
    backup_configuration { enabled = true }
    ip_configuration { ipv4_enabled = true }
  }
}

resource "google_redis_instance" "cache" {
  name           = "lispflow-${var.environment}"
  tier           = "BASIC"
  memory_size_gb = var.redis_memory_gb
  region         = var.region
}

resource "google_storage_bucket" "state" {
  name     = "${var.project_id}-tfstate-${var.environment}"
  location = var.region
  versioning { enabled = true }
}

output "cluster_endpoint" { value = google_container_cluster.primary.endpoint }
output "db_connection" { value = google_sql_database_instance.postgres.connection_name }
output "redis_host" { value = google_redis_instance.cache.host }
