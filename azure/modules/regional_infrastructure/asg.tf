# Render a part using a `template_file`
data "template_file" "script" {
  template = "${file("${path.module}/files/init.sh.tpl")}"

  vars = {
    prefix               = var.prefix
    chain                = var.chain
    total_instance_count = var.total_instance_count
    lb-primary           = cidrhost(var.subnet_cidrs[0],10)
    lb-secondary         = cidrhost(var.subnet_cidrs[1],10)
    lb-tertiary          = cidrhost(var.subnet_cidrs[2],10)
  }
}

# Render a multi-part cloud-init config making use of the part
# above, and other source files
data "template_cloudinit_config" "config" {
  gzip          = true
  base64_encode = true

  # Main cloud-config configuration file.
  part {
    filename     = "init.cfg"
    content_type = "text/cloud-config"
    content      = file("${path.module}/files/config.yaml")
  }

  part {
    content_type = "text/x-shellscript"
    content      = data.template_file.script.rendered
  }
}

resource "azurerm_linux_virtual_machine_scale_set" "polkadot" {
  name                 = "${var.prefix}-instance-${var.region_prefix}"
  computer_name_prefix = var.region_prefix
  resource_group_name  = var.rg
  location             = var.region
  sku                  = var.instance_type
  instances            = var.instance_count
  custom_data          = data.template_cloudinit_config.config.rendered
  health_probe_id      = length(module.private_lb.azurerm_lb_hc) > 0 ? module.private_lb.azurerm_lb_hc[0] : null

  automatic_instance_repair {
    enabled = true
  }
  

  identity {
    type = "SystemAssigned"
  }

  admin_ssh_key {
    username   = "polkadot"
    public_key = var.key_content
  }

  source_image_reference {
    publisher = "OpenLogic"
    offer     = "CentOS"
    sku       = "7-CI"
    version   = "latest"
  }

  os_disk {
    storage_account_type = var.sa_type
    caching              = "ReadWrite"
  }

  data_disk {
    storage_account_type = var.sa_type
    disk_size_gb         = var.disk_size
    caching              = "ReadWrite"
    lun                  = 0
  }

  network_interface {
    name    = "${var.prefix}-ni-${var.region_prefix}"
    primary = true

    ip_configuration {
      load_balancer_backend_address_pool_ids = [module.private_lb.azurerm_lb_backend_address_pool_id]
      name      = "primary-${var.region_prefix}"
      primary   = true
      public_ip_address {
        name = "public-${var.region_prefix}"
      }
      subnet_id = azurerm_subnet.polkadot.id
    }
  }

  admin_username = "polkadot"

  tags = {
    prefix = var.prefix
  }

  depends_on = [module.private_lb]
}
