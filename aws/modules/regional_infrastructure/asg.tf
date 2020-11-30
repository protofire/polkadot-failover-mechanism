data "aws_ami" "amazon-linux-2" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["amzn2-ami-hvm*"]
  }
}

resource "aws_launch_template" "polkadot" {
  name_prefix                          = "${var.prefix}-polkadot-validator"
  image_id                             = data.aws_ami.amazon-linux-2.id
  instance_initiated_shutdown_behavior = "terminate"
  instance_type                        = var.instance_type
  key_name                             = var.key_content == "" ? var.key_name : "${var.prefix}-${var.key_name}"

  monitoring {
    enabled = true
  }

  iam_instance_profile {
    name = var.asg_role
  }

  network_interfaces {
    delete_on_termination       = true
    subnet_id                   = var.subnet.id
    security_groups             = [aws_security_group.validator-node.id]
    associate_public_ip_address = var.expose_ssh || var.expose_prometheus
  }

  tag_specifications {
    resource_type = "instance"

    tags = {
      Name   = "${var.prefix}-failover-validator"
      prefix = var.prefix
      region = var.region
    }
  }

  tag_specifications {
    resource_type = "volume"

    tags = {
      Name   = "${var.prefix}-failover-validator-root"
      prefix = var.prefix
      region = var.region
    }
  }

  user_data = base64encode(templatefile("${path.module}/files/init.sh.tpl", {
    prefix                   = var.prefix,
    primary-region           = var.regions[0],
    secondary-region         = var.regions[1],
    tertiary-region          = var.regions[2],
    autoscaling-name         = "${var.prefix}-polkadot-validator-${var.region_prefix}"
    chain                    = var.chain,
    disk_size                = var.disk_size,
    cpu_limit                = var.cpu_limit,
    ram_limit                = var.ram_limit,
    lb-primary               = var.instance_count_primary > 0 ? var.lbs[0].dns_name : "",
    lb-secondary             = var.instance_count_secondary > 0 ? var.lbs[1].dns_name : "",
    lb-tertiary              = var.instance_count_tertiary > 0 ? var.lbs[2].dns_name : "",
    delete_on_termination    = var.delete_on_termination,
    total_instance_count     = var.total_instance_count,
    docker_image             = var.docker_image,
    prometheus_port          = var.prometheus_port,
    polkadot_prometheus_port = local.polkadot_prometheus_port,
    expose_prometheus        = var.expose_prometheus,
  }))

}

resource "aws_autoscaling_group" "polkadot" {
  name                      = "${var.prefix}-polkadot-validator-${var.region_prefix}"
  max_size                  = var.instance_count
  min_size                  = var.instance_count
  desired_capacity          = var.instance_count
  wait_for_elb_capacity     = var.instance_count
  wait_for_capacity_timeout = "30m"
  default_cooldown          = 300
  health_check_type         = "ELB"
  health_check_grace_period = 300

  launch_template {
    id      = aws_launch_template.polkadot.id
    version = aws_launch_template.polkadot.latest_version
  }

  target_group_arns = flatten([
    aws_lb_target_group.http.id,
    aws_lb_target_group.dns.id,
    aws_lb_target_group.rpc.id,
    aws_lb_target_group.lan.id,
    aws_lb_target_group.wan.id,
    aws_lb_target_group.polkadot.id,
    length(aws_lb_target_group.prometheus) > 0 ? [aws_lb_target_group.prometheus[0].id] : []
  ])

  vpc_zone_identifier = [var.subnet.id]

  depends_on = [
    aws_ssm_parameter.keys,
    aws_ssm_parameter.seeds,
    aws_ssm_parameter.types,
    aws_ssm_parameter.name,
    aws_ssm_parameter.cpu_limit,
    aws_ssm_parameter.ram_limit
  ]

}
