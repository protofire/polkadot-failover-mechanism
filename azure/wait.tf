resource "null_resource" "primary_port" {
  depends_on = [module.primary_region.scale_set_id]
  count      = var.wait_vmss ? 1 : 0
  provisioner "local-exec" {
    command = "bash ${path.module}/../init-helpers/azure/ports.sh \"${local.p2p_port}\" \"${var.azure_rg}\" \"${module.primary_region.scale_set_name}\""
  }
}

resource "null_resource" "secondary_port" {
  depends_on = [module.secondary_region.scale_set_id]
  count      = var.wait_vmss ? 1 : 0
  provisioner "local-exec" {
    command = "bash ${path.module}/../init-helpers/azure/ports.sh \"${local.p2p_port}\" \"${var.azure_rg}\" \"${module.secondary_region.scale_set_name}\""
  }
}

resource "null_resource" "tertiary_port" {
  depends_on = [module.tertiary_region.scale_set_id]
  count      = var.wait_vmss ? 1 : 0
  provisioner "local-exec" {
    command = "bash ${path.module}/../init-helpers/azure/ports.sh \"${local.p2p_port}\" \"${var.azure_rg}\" \"${module.tertiary_region.scale_set_name}\""
  }
}
