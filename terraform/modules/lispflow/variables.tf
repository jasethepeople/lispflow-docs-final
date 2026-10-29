variable "environment" { type = string }
variable "project_id" { type = string default = "lispflow" }
variable "region" { type = string default = "us-central1" }
variable "cluster_name" { type = string default = "lispflow" }
variable "node_count" { type = number default = 3 }
variable "machine_type" { type = string default = "e2-standard-4" }
variable "db_tier" { type = string default = "db-g1-small" }
variable "redis_memory_gb" { type = number default = 1 }
variable "image_repository" { type = string default = "ghcr.io/livininthelast/lispflow" }
variable "image_tag" { type = string default = "latest" }
