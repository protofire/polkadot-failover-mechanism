locals {
  instance_ids = concat(
    data.aws_instances.ec2-asg-instances-info-1.ids,
    data.aws_instances.ec2-asg-instances-info-2.ids,
    data.aws_instances.ec2-asg-instances-info-3.ids,
  )
}