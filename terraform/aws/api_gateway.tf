resource "aws_api_gateway_rest_api" "main" {
  count = var.use_api_gateway ? 1 : 0
  name  = var.api_gateway_name
}

resource "aws_api_gateway_stage" "main" {
  count         = var.use_api_gateway ? 1 : 0
  deployment_id = aws_api_gateway_deployment.main[0].id
  rest_api_id   = aws_api_gateway_rest_api.main[0].id
  stage_name    = "main"
}

resource "aws_api_gateway_resource" "main" {
  count       = var.use_api_gateway ? 1 : 0
  path_part   = "webhook"
  parent_id   = aws_api_gateway_rest_api.main[0].root_resource_id
  rest_api_id = aws_api_gateway_rest_api.main[0].id
}

resource "aws_api_gateway_method" "main" {
  count         = var.use_api_gateway ? 1 : 0
  rest_api_id   = aws_api_gateway_rest_api.main[0].id
  resource_id   = aws_api_gateway_resource.main[0].id
  http_method   = "POST"
  authorization = "NONE"
}

resource "aws_api_gateway_integration" "main" {
  count                   = var.use_api_gateway ? 1 : 0
  rest_api_id             = aws_api_gateway_rest_api.main[0].id
  resource_id             = aws_api_gateway_resource.main[0].id
  http_method             = aws_api_gateway_method.main[0].http_method
  integration_http_method = "POST"
  type                    = "AWS_PROXY"
  uri                     = aws_lambda_function.main.invoke_arn
}

resource "aws_api_gateway_deployment" "main" {
  count       = var.use_api_gateway ? 1 : 0
  rest_api_id = aws_api_gateway_rest_api.main[0].id

  triggers = {
    redeployment = sha1(jsonencode([
      aws_api_gateway_resource.main[0].id,
      aws_api_gateway_method.main[0].id,
      aws_api_gateway_integration.main[0].id,
    ]))
  }

  lifecycle {
    create_before_destroy = true
  }
}

resource "aws_lambda_permission" "main" {
  count         = var.use_api_gateway ? 1 : 0
  statement_id  = "AllowLambuildInvoke"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.main.function_name
  principal     = "apigateway.amazonaws.com"

  # More: http://docs.aws.amazon.com/apigateway/latest/developerguide/api-gateway-control-access-using-iam-policies-to-invoke-api.html
  source_arn = "arn:aws:execute-api:${data.aws_region.current.region}:${data.aws_caller_identity.current.account_id}:${aws_api_gateway_rest_api.main[0].id}/*/${aws_api_gateway_method.main[0].http_method}${aws_api_gateway_resource.main[0].path}"
}

data "aws_caller_identity" "current" {}
data "aws_region" "current" {}
