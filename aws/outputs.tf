output "lbs" {
  value = [module.primary_network.lb.arn, module.secondary_network.lb.arn, module.tertiary_network.lb.arn]
}

