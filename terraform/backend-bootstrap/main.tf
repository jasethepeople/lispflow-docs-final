resource "google_storage_bucket" "tfstate" {
  for_each = toset(["staging","prod"])
  name     = "lispflow-tfstate-${each.key}"
  location = "US"
  versioning { enabled = true }
  lifecycle { prevent_destroy = true }
}
