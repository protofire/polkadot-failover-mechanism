resource "aws_ssm_parameter" "keys" {
  for_each = var.validator_keys

  name        = "/polkadot/validator-failover/${var.prefix}/keys/${each.key}/key"
  description = "Validator key"
  type        = "String"
  value       = each.value.key

  tags = {
    environment = var.prefix
    type        = "key"
  }
}

resource "aws_ssm_parameter" "seeds" {
  for_each = var.validator_keys

  name        = "/polkadot/validator-failover/${var.prefix}/keys/${each.key}/seed"
  description = "Validator private seed"
  type        = "SecureString"
  value       = each.value.seed

  tags = {
    environment = var.prefix
    type        = "seed"
  }
}

resource "aws_ssm_parameter" "types" {
  for_each = var.validator_keys

  name        = "/polkadot/validator-failover/${var.prefix}/keys/${each.key}/type"
  description = "Validator key type"
  type        = "String"
  value       = each.value.type

  tags = {
    environment = var.prefix
    type        = "type"
  }
}

resource "aws_ssm_parameter" "name" {
  name        = "/polkadot/validator-failover/${var.prefix}/name"
  description = "Validator name"
  type        = "String"
  value       = var.validator_name

  tags = {
    environment = var.prefix
  }
}

resource "aws_ssm_parameter" "ram_limit" {
  name        = "/polkadot/validator-failover/${var.prefix}/ram_limit"
  description = "Validator key"
  type        = "String"
  value       = var.ram_limit

  tags = {
    environment = var.prefix
  }
}

resource "aws_ssm_parameter" "cpu_limit" {
  name        = "/polkadot/validator-failover/${var.prefix}/cpu_limit"
  description = "Validator key"
  type        = "String"
  value       = var.cpu_limit

  tags = {
    environment = var.prefix
  }
}

resource "aws_ssm_parameter" "node_key" {
  name        = "/polkadot/validator-failover/${var.prefix}/node_key"
  description = "Node ID key"
  type        = "String"
  value       = var.node_key

  tags = {
    environment = var.prefix
  }
}
