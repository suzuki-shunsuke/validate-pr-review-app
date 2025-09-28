locals {
  config            = yamldecode(file("${path.module}/config.yaml"))
  use_api_gateway   = lookup(local.config.aws, "use_api_gateway", false)
  api_gateway_count = local.use_api_gateway ? 1 : 0
}
