variable "environment" {
  description = "Environment tag, e.g prod"
}

variable "name" {
  description = "Name of this siteshield instance, e.g. 'shop'"
}

variable "security_group_id" {
  description = "Security group id which will be used by the siteshield."
}

variable "akamai_ssid" {
  description = "Akamai SSID."
}

variable "akamai_edgegrid_host" {
  description = "Akamai edgegrid host."
}

variable "akamai_edgegrid_client_token" {
  description = "Akamai edgegrid client token."
}

variable "akamai_edgegrid_client_secret" {
  description = "Akamai edgegrid client secret."
}

variable "akamai_edgegrid_access_token" {
  description = "Akamai edgegrid access token."
}

// ******************************************************************
//
// ******************************************************************

data "external" "lambda" {
  program = ["bash", "${path.module}/update.sh", "${path.module}"]
}

// ******************************************************************
// Policy
// ******************************************************************

resource "aws_iam_role" "main" {
  name               = "siteshield-${var.name}-${var.environment}-role"
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "lambda.amazonaws.com"
      },
      "Effect": "Allow"
    }
  ]
}
EOF
}

resource "aws_iam_policy" "main" {
  name   = "siteshield-${var.name}-${var.environment}-policy"
  policy = <<POLICY
{
    "Version": "2012-10-17",
    "Statement": [
        {
          "Effect": "Allow",
          "Action": [
            "kms:Decrypt",
            "kms:DescribeKey",
            "kms:GetKeyPolicy",
            "ssm:GetParameters"
          ],
          "Resource": "*"
        },
        {
            "Effect": "Allow",
            "Action": [
                "logs:CreateLogGroup",
                "logs:CreateLogStream",
                "logs:PutLogEvents"
            ],
            "Resource": "arn:aws:logs:*:*:*"
        },
        {
            "Effect": "Allow",
            "Action": [
                "ec2:DescribeSecurityGroups",
                "ec2:RevokeSecurityGroupIngress",
                "ec2:AuthorizeSecurityGroupIngress"
            ],
            "Resource": "*"
        }
    ]
}
POLICY
}

resource "aws_iam_role_policy_attachment" "main" {
  role       = "${aws_iam_role.main.id}"
  policy_arn = "${aws_iam_policy.main.id}"
}

// ******************************************************************
// Function
// ******************************************************************

resource "aws_lambda_function" "main" {
  filename         = "${path.module}/main.zip"
  source_code_hash = "${base64sha256(file("${path.module}/${data.external.lambda.result.filename}"))}"
  function_name    = "siteshield-${var.name}-${var.environment}"
  handler          = "main"
  runtime          = "go1.x"
  role             = "${aws_iam_role.main.arn}"
  timeout          = "10"

  environment {
    variables = {
      AWS_SECGROUP_ID = "${var.security_group_id}"
      KMS_AKAMAI_SSID = "${var.akamai_ssid}"
      KMS_AKAMAI_EDGEGRID_HOST          = "${var.akamai_edgegrid_host}"
      KMS_AKAMAI_EDGEGRID_CLIENT_TOKEN  = "${var.akamai_edgegrid_client_token}"
      KMS_AKAMAI_EDGEGRID_CLIENT_SECRET = "${var.akamai_edgegrid_client_secret}"
      KMS_AKAMAI_EDGEGRID_ACCESS_TOKEN  = "${var.akamai_edgegrid_access_token}"
      CSS_ARGS = "-add-missing,-delete-obsolete"
    }
  }

  depends_on = ["data.external.lambda"]
}

// ******************************************************************
// Cloudwatch action (Cron like trigger)
// ******************************************************************

resource "aws_cloudwatch_event_rule" "main" {
  name                = "siteshield-${var.name}-${var.environment}"
  description         = "Fires every 6 hours."
  schedule_expression = "rate(6 hours)"
}

resource "aws_cloudwatch_event_target" "main" {
  rule  = "${aws_cloudwatch_event_rule.main.name}"
  arn   = "${aws_lambda_function.main.arn}"
  depends_on = ["aws_cloudwatch_event_rule.main"]
}

resource "aws_lambda_permission" "main" {
  statement_id  = "AllowExecutionFromCloudWatch"
  action        = "lambda:InvokeFunction"
  function_name = "${aws_lambda_function.main.function_name}"
  principal     = "events.amazonaws.com"
  source_arn    = "${aws_cloudwatch_event_rule.main.arn}"
}