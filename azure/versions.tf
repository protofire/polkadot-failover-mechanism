terraform {
  required_version = ">= 0.12"

  backend "azurerm" {
    version = "~> 2.10"
  }

}
