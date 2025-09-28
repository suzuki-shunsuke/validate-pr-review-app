resource "aws_lambda_function_url" "main" {
  count              = local.use_api_gateway ? 0 : 1
  function_name      = aws_lambda_function.main.function_name
  authorization_type = "NONE"
}
