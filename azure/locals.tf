locals {
  key_vault_name    = "${var.prefix}-polkadot-vault"
  p2p_port          = 30333
  metrics_namespace = "${var.prefix}/validator"
}
