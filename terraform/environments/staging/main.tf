terraform {
  backend "gcs" { bucket = "lispflow-tfstate-staging" prefix = "terraform/state" }
}
module "lispflow" {
  source = "../../modules/lispflow"
  environment = "staging"
  project_id  = "lispflow-staging"
  node_count  = 2
  machine_type = "e2-medium"
}
