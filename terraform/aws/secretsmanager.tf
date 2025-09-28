resource "aws_secretsmanager_secret" "main" {
  name = yamldecode(file("${path.module}/config.yaml")).aws.secret_id
}

resource "aws_secretsmanager_secret_version" "main" {
  secret_id                = aws_secretsmanager_secret.main.id
  secret_string_wo         = jsonencode(yamldecode(file("${path.module}/secret.yaml")))
  secret_string_wo_version = 1
}

resource "aws_iam_role_policy" "read_secret" {
  name   = "read-secret"
  policy = data.aws_iam_policy_document.read_secret.json
  role   = aws_iam_role.lambda.name
}

data "aws_iam_policy_document" "read_secret" {
  statement {
    actions = ["secretsmanager:GetSecretValue"]
    resources = [
      aws_secretsmanager_secret_version.main.arn,
    ]
  }
}
