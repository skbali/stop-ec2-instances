provider "aws" {
  region  = "us-east-1"
  profile = "default"
}

data "archive_file" "stop_ec2_zip" {
  type        = "zip"
  output_path = "./stop_ec2/stop_ec2.zip"

  source_file = "./stop_ec2/code/build/bootstrap"
}

resource "aws_lambda_function" "stop_ec2" {
  function_name = "stop-ec2-go"
  handler       = "bootstrap"
  role          = aws_iam_role.stop_ec2_go_lambda_role.arn

  runtime          = "provided.al2"
  timeout          = 120
  memory_size      = 128
  filename         = data.archive_file.stop_ec2_zip.output_path
  source_code_hash = data.archive_file.stop_ec2_zip.output_base64sha256

  environment {
    variables = {
      REGION    = var.region
      TOPIC_ARN = data.aws_sns_topic.site_monitor_sns_topic.arn
      MAX_HOURS = var.max_hours
    }
  }
  tags = var.tags
}

#resource "aws_cloudwatch_event_rule" "every_30_min" {
#  name                = "stop-ec2-go-30-min"
#  schedule_expression = "rate(30 minutes)"
#}
#
#resource "aws_cloudwatch_event_target" "stop_ec2_go_lambda_target" {
#  arn  = aws_lambda_function.stop_ec2.arn
#  rule = aws_cloudwatch_event_rule.every_30_min.name
#}

#resource "aws_lambda_permission" "stop_ec2_go_lamabda_perms" {
#  action        = "lambda:InvokeFunction"
#  function_name = aws_lambda_function.stop_ec2.function_name
#  principal     = "events.amazonaws.com"
#  source_arn    = aws_cloudwatch_event_rule.every_30_min.arn
#}

data "aws_sns_topic" "site_monitor_sns_topic" {
  name = "site-monitor"
}


