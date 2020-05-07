# Key must be supplied at the init stage
# https://github.com/hashicorp/terraform/issues/13022

terraform {
  backend "s3" {
  }
}
