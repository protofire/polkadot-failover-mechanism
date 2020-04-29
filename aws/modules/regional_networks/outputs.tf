output "lb" {
  value = aws_lb.polkadot
}

output "vpc" {
  value = aws_vpc.vpc
}

output "subnet" {
  value = aws_subnet.default
}