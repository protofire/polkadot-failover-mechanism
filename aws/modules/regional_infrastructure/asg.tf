data "aws_ami" "amazon-linux-2" {
  most_recent = true
  owners = [
  "amazon"]

  filter {
    name = "name"
    values = [
    "amzn2-ami-hvm*"]
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
    delete_on_termination = true
    subnet_id             = var.subnet.id
    security_groups       = [aws_security_group.validator-node.id]
  }

  tag_specifications {
    resource_type = "instance"

    tags = {
      Name   = "${var.prefix}-failover-validator"
      prefix = var.prefix
    }
  }

  tag_specifications {
    resource_type = "volume"

    tags = {
      Name   = "${var.prefix}-failover-validator-root"
      prefix = var.prefix
    }
  }

  user_data = base64encode(templatefile("${path.module}/files/init.sh.tpl", {
    prefix                = var.prefix,
    primary-region        = var.regions[0],
    secondary-region      = var.regions[1],
    tertiary-region       = var.regions[2],
    autoscaling-name      = "${var.prefix}-polkadot-validator",
    chain                 = var.chain,
    disk_size             = var.disk_size,
    cpu_limit             = var.cpu_limit,
    ram_limit             = var.ram_limit,
    lb-primary            = var.lbs[0].dns_name,
    lb-secondary          = var.lbs[1].dns_name,
    lb-tertiary           = var.lbs[2].dns_name,
    delete_on_termination = var.delete_on_termination,
    total_instance_count  = var.total_instance_count,
    docker_image          = var.docker_image,
  }))
}

resource "aws_autoscaling_group" "polkadot" {
  name                      = "${var.prefix}-polkadot-validator"
  max_size                  = var.instance_count
  min_size                  = var.instance_count
  desired_capacity          = var.instance_count
  wait_for_elb_capacity     = var.instance_count
  default_cooldown          = 300
  health_check_type         = "ELB"
  health_check_grace_period = 300

  launch_template {
    id      = aws_launch_template.polkadot.id
    version = "$Latest"
  }

  target_group_arns = [
    aws_lb_target_group.http.id,
    aws_lb_target_group.dns.id,
    aws_lb_target_group.rpc.id,
    aws_lb_target_group.lan.id,
    aws_lb_target_group.wan.id,
  aws_lb_target_group.polkadot.id]
  vpc_zone_identifier = [
  var.subnet.id]

  depends_on = [
    aws_ssm_parameter.keys,
    aws_ssm_parameter.seeds,
    aws_ssm_parameter.types,
    aws_ssm_parameter.name,
    aws_ssm_parameter.cpu_limit,
  aws_ssm_parameter.ram_limit]
}
