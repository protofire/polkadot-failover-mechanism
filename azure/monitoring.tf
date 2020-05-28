resource "azurerm_monitor_action_group" "main" {
  name                = "${var.prefix}-ag"
  resource_group_name = var.azure_rg
  short_name          = var.prefix
 
  tags = {
    prefix = var.prefix
  }
  
  email_receiver {
    name                    = "Admin"
    email_address           = var.admin_email
    use_common_alert_schema = true
  }
}
