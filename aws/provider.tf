provider "aws" {
  version    = "~> 2.67"
  profile    = length(var.aws_profiles) > 0 ? var.aws_profiles[0] : null
  access_key = length(var.aws_access_keys) > 0 ? var.aws_access_keys[0] : null
  secret_key = length(var.aws_secret_keys) > 0 ? var.aws_secret_keys[0] : null
  region     = var.aws_regions[0]
  alias      = "primary"

}

provider "aws" {
  version    = "~> 2.67"
  profile    = length(var.aws_profiles) > 0 ? element(var.aws_profiles, 1) : null
  access_key = length(var.aws_access_keys) > 0 ? element(var.aws_access_keys, 1) : null
  secret_key = length(var.aws_secret_keys) > 0 ? element(var.aws_secret_keys, 1) : null
  region     = var.aws_regions[1]
  alias      = "secondary"
}

provider "aws" {
  version    = "~> 2.67"
  profile    = length(var.aws_profiles) > 0 ? element(var.aws_profiles, 2) : null
  access_key = length(var.aws_access_keys) > 0 ? element(var.aws_access_keys, 2) : null
  secret_key = length(var.aws_secret_keys) > 0 ? element(var.aws_secret_keys, 2) : null
  region     = var.aws_regions[2]
  alias      = "tertiary"
}
