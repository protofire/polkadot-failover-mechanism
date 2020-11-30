resource "aws_instance" "prometheus" {
  ami                    = data.aws_ami.amazon-linux-2.id
  instance_type          = var.instance_type
  key_name               = var.key_name != "" ? var.key_name : null
  monitoring             = true
  vpc_security_group_ids = [aws_security_group.prometheus.id]
  subnet_id              = var.subnet_id

  user_data_base64 = base64encode(templatefile("${path.module}/files/init.sh.tpl", {
    url_primary     = var.instance_count_primary > 0 ? "http://${var.lbs[0].dns_name}:${var.prometheus_port}${var.metrics_path}" : "",
    url_secondary   = var.instance_count_secondary > 0 ? "http://${var.lbs[1].dns_name}:${var.prometheus_port}${var.metrics_path}" : "",
    url_tertiary    = var.instance_count_tertiary > 0 ? "http://${var.lbs[2].dns_name}:${var.prometheus_port}${var.metrics_path}" : "",
    prometheus_port = var.prometheus_port,
  }))

  tags = {
    Name = "prometheus"
  }
}

resource "aws_eip" "prometheus" {
  instance = aws_instance.prometheus.id
  vpc      = true
}
