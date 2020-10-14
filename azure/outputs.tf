output "vnet" {
  value = [
    module.primary_region.vnet,
    module.secondary_region.vnet,
    module.tertiary_region.vnet,
  ]
}

output "subnet" {
  value = [
    module.primary_region.subnet,
    module.secondary_region.subnet,
    module.tertiary_region.subnet,
  ]
}

output "principal_id" {
  value = [
    module.primary_region.principal_id,
    module.secondary_region.principal_id,
    module.tertiary_region.principal_id,
  ]
}

output "scale_set_id" {
  value = [
    module.primary_region.scale_set_id,
    module.secondary_region.scale_set_id,
    module.tertiary_region.scale_set_id,
  ]
}

output "private_lb_id" {
  value = [
    module.primary_region.private_lb_id,
    module.secondary_region.private_lb_id,
    module.tertiary_region.private_lb_id,
  ]
}
