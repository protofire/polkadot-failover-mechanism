locals {

  lb_address = {
    primary   = cidrhost(var.subnet_cidrs[0], 10)
    secondary = cidrhost(var.subnet_cidrs[1], 10)
    tertiary  = cidrhost(var.subnet_cidrs[2], 10)
  }
  private_ip_address = cidrhost(var.subnet_cidrs[0], 25)
}
