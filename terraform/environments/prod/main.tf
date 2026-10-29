terraform {
  backend "gcs" { bucket = "lispflow-tfstate-prod" prefix = "terraform/state" }
}
module "lispflow" {
  source = "../../modules/lispflow"
  environment = "prod"
  project_id  = "lispflow-prod"
  node_count  = 3
  machine_type = "e2-standard-4"
  db_tier = "db-custom-2-4096"
}
