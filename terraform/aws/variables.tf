variable "region" {
  type    = string
  default = "us-east-1"
}

variable "secretsmanager_secret_name_main" {
  type    = string
  default = "validate-pr-review-app"
}

variable "lambda_architecture" {
  type        = string
  description = "Lambda Architecture"
  default     = "arm64"
}

variable "zip_path" {
  type        = string
  description = "Lambda Zip File Path"
  default     = "validate-pr-review-app_lambda_linux_arm64.zip"
}

variable "function_name" {
  type        = string
  description = "Lambda Function Name"
  default     = "validate-pr-review-app"
}

variable "lambda_role_name" {
  type        = string
  description = "Lambda Role Name"
  default     = "validate-pr-review-app"
}

variable "lambda_role_path" {
  type        = string
  description = "Lambda Role Path"
  default     = "/service-role/"
}
