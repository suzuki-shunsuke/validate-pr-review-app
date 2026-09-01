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
- `github_app_private_key`: GitHub App Private Key.

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
