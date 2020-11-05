terraform {
  required_version = ">= 0.13"
  required_providers {
    polkadot = {
      versions = ["0.1"]
      source   = "polkadot-failover-mechanism/aws/polkadot"
    }
  }
}