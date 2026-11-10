# Secrets

## Sources

The following secret sources are supported:

1. Environment Variable `SECRET`
2. File. The file path is specified by the environment variable `SECRET_FILE`
3. AWS Secrets Manager

e.g.

```yaml
aws:
  secret_id: validate-pr-review-app # Secret ID in AWS Secrets Manager
```

4. [Google Secret Manager](https://cloud.google.com/security/products/secret-manager)

e.g.

```yaml
google_cloud:
  secret_name: validate-pr-review-app
```

## Secrets

- `webhook_secret`: [GitHub App's Webhook Secret. This is used to validate webhook](https://docs.github.com/en/webhooks/using-webhooks/validating-webhook-deliveries)
- `github_app_private_key`: GitHub App Private Key. This isn't needed and must not be set when the private key is managed with AWS KMS.

e.g.

```yaml
webhook_secret: 0123456789abcdefghijklmn
github_app_private_key: |
  -----BEGIN RSA PRIVATE KEY-----
  ...
```

> [!WARNING]
> When using AWS Secrets Manager Web UI, multi-line values are not supported.
> You should convert the private key and webhook secret into JSON before storing.

## Manage the GitHub App private key with AWS KMS

Instead of storing the GitHub App private key as a secret, you can import it into AWS KMS and delegate signing JSON Web Tokens to AWS KMS.
validate-pr-review-app never reads the private key, so the private key can't leak from it.

```yaml
aws:
  kms_key_id: arn:aws:kms:us-east-1:123456789012:key/00000000-0000-0000-0000-000000000000
```

- `kms_key_id` accepts a key ID, key ARN, alias name, or alias ARN
- If `kms_key_id` is an ARN, the region of the ARN is used. Otherwise, the region is resolved by the AWS SDK, such as the environment variable `AWS_REGION`. A key ARN works even if the key is in a different region from `AWS_REGION`
- `secret_id` is optional. You can pass `webhook_secret` with an environment variable or a file instead of AWS Secrets Manager
- `github_app_private_key` must not be set
- `aws` and `google_cloud` can't be set together, so this feature isn't available with Google Secret Manager

### AWS KMS key

The key must be created with the key spec `RSA_2048`, the key usage `SIGN_VERIFY`, and the origin `EXTERNAL`, and then the GitHub App private key is imported into it.
Please remove the private key from your machine after importing it.

- https://github.com/suzuki-shunsuke/create-github-app-token-aws-kms#the-kms-key
- https://github.com/suzuki-shunsuke/agent-skills/blob/main/skills/github-app-private-key-aws-kms/SKILL.md

### IAM policy

The IAM role of validate-pr-review-app (e.g. the Lambda execution role, or the IAM role for the Kubernetes Pod or EC2 instance) needs only `kms:Sign` on the key.

```json
{
  "Effect": "Allow",
  "Action": "kms:Sign",
  "Resource": "arn:aws:kms:us-east-1:123456789012:key/00000000-0000-0000-0000-000000000000"
}
```

> [!WARNING]
> Specify the key ARN in `Resource`, even if `kms_key_id` is an alias.
> An alias ARN in `Resource` doesn't allow `kms:Sign` on the key.
> If you want to restrict access by the alias, use the condition key `kms:RequestAlias`.

### Troubleshooting

The AWS KMS key isn't checked at startup.
Misconfigurations such as a wrong key ID, a missing `kms:Sign` permission, or a wrong region are detected when validate-pr-review-app processes a webhook.
Webhooks are processed in the background after responding to GitHub, so such errors don't appear as failed webhook deliveries in GitHub.
No check run is created, and the error is output only to the logs.
If a check run isn't created after you enable AWS KMS, please check the logs.
