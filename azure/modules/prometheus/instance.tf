data "template_file" "script" {
  template = file("${path.module}/files/init.sh.tpl")

  vars = {
    prefix          = var.prefix
    url_primary     = var.instance_count_primary > 0 ? "http://${local.lb_address.primary}:${var.prometheus_port}${var.metrics_path}" : ""
    url_secondary   = var.instance_count_secondary > 0 ? "http://${local.lb_address.secondary}:${var.prometheus_port}${var.metrics_path}" : ""
    url_tertiary    = var.instance_count_tertiary > 0 ? "http://${local.lb_address.tertiary}:${var.prometheus_port}${var.metrics_path}" : ""
    prometheus_port = var.prometheus_port
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

resource "azurerm_public_ip" "prometheus" {
  name                = "prometheus"
  location            = var.region
  resource_group_name = var.rg
  allocation_method   = "Static"

  tags = {
    prefix = var.prefix
  }
}

resource "azurerm_network_interface" "prometheus" {
  name                = "prometheus"
  location            = var.region
  resource_group_name = var.rg

  ip_configuration {
    name                          = "prometheus-ip"
    subnet_id                     = var.subnet_id
    private_ip_address_allocation = "Static"
    private_ip_address            = local.private_ip_address
    public_ip_address_id          = azurerm_public_ip.prometheus.id
  }
}

resource "azurerm_linux_virtual_machine" "prometheus" {
  name                            = "${var.prefix}-instance-prometheus"
  resource_group_name             = var.rg
  location                        = var.region
  size                            = var.instance_type
  network_interface_ids           = [azurerm_network_interface.prometheus.id]
  disable_password_authentication = true
  admin_username                  = var.admin_user
  custom_data                     = data.template_cloudinit_config.config.rendered

  admin_ssh_key {
    username   = var.ssh_user
    public_key = var.ssh_key_content
  }

  os_disk {
    storage_account_type = var.sa_type
    caching              = "ReadWrite"
  }

  source_image_reference {
    publisher = "OpenLogic"
    offer     = "CentOS"
    sku       = "7.7"
    version   = "latest"
  }

  tags = {
    prefix = var.prefix
  }

}
