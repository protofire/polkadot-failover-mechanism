# MONITORING SECURITY
resource "aws_iam_role_policy_attachment" "monitoring" {
  provider = aws.primary

  role       = aws_iam_role.monitoring.name
  policy_arn = aws_iam_policy.monitoring.arn
}

resource "aws_iam_policy" "monitoring" {
  provider = aws.primary

  name        = "${var.prefix}-polkadot-validator"
  description = "This is the policy that allows EC2 to send metrics to the CloudWatch"

  # AWS bug - there is no option to limit ec2:ModifyInstanceAttribute parameter. Should be fixed in later releases
  policy = <<EOF
{
    "Version": "2012-10-17",
    "Statement": [
        {
         "Effect":"Allow",
         "Action": [
                "ssm:GetParametersByPath",
                "ssm:GetParameters",
                "ssm:GetParameter"
            ],
         "Resource":"arn:aws:ssm:*:*:parameter/polkadot/validator-failover/${var.prefix}/*"
        },
        {
            "Effect": "Allow",
            "Action": "ec2:CreateVolume",
            "Resource": "arn:aws:ec2:*:*:volume/*",
            "Condition": {
                "StringEquals": {
                    "aws:RequestTag/Name": "${var.prefix}-polkadot-failover-data",
                    "aws:RequestTag/prefix": "${var.prefix}"
                },
                "ForAllValues:StringEquals": {
                    "aws:TagKeys": [
                        "prefix",
                        "Name"
                    ]
                }
            }
        },
        {
            "Effect": "Allow",
            "Action": "ec2:AttachVolume",
            "Resource": [
                "arn:aws:ec2:*:*:instance/*",
                "arn:aws:ec2:*:*:volume/*"
            ],
            "Condition": {
                "StringEquals": {
                    "ec2:ResourceTag/prefix": "${var.prefix}"
                },
                "ForAllValues:StringEquals": {
                    "aws:TagKeys": [
                        "prefix",
                        "Name"
                    ]
                },
                "ForAnyValue:StringEquals": {
                    "ec2:ResourceTag/Name": [
                        "${var.prefix}-polkadot-failover-data",
                        "${var.prefix}-failover-validator"
                    ]
                }
            }
        },
        {
            "Effect": "Allow",
            "Action": "ec2:CreateTags",
            "Resource": "arn:aws:ec2:*:*:volume/*",
            "Condition": {
                "StringEquals": {
                    "aws:RequestTag/Name": "${var.prefix}-polkadot-failover-data",
                    "ec2:CreateAction": "CreateVolume",
                    "aws:RequestTag/prefix": "${var.prefix}"
                },
                "ForAllValues:StringEquals": {
                    "aws:TagKeys": [
                        "prefix",
                        "Name"
                    ]
                }
            }
        },
        {
            "Effect": "Allow",
            "Action": [
                "cloudwatch:PutMetricData",
                "ec2:DescribeVolumes",
                "ec2:DescribeInstances",
                "ec2:DescribeTags",
                "autoscaling:DescribeAutoScalingGroups",
                "ec2:ModifyInstanceAttribute"
            ],
            "Resource": "*"
        }      
    ]
}
EOF
}

resource "aws_iam_instance_profile" "monitoring" {
  provider = aws.primary

  name = "${var.prefix}-polkadot.validator"
  role = aws_iam_role.monitoring.name

  depends_on = [aws_iam_policy.monitoring]
}

resource "aws_iam_role" "monitoring" {
  provider = aws.primary

  name = "${var.prefix}-polkadot-validator"
  path = "/"

  assume_role_policy = <<EOF
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Action": "sts:AssumeRole",
            "Principal": {
               "Service": "ec2.amazonaws.com"
            },
            "Effect": "Allow",
            "Sid": ""
        }
    ]
}
EOF
}


