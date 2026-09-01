# Environment Variables

Either `CONFIG` or `CONFIG_FILE` is required.

- `CONFIG`: A YAML string for configuration
- `CONFIG_FILE`: A configuration file path

[See Configuration for the content of `CONFIG` and `CONFIG_FILE`.](config.md)

For HTTP Server:

- `PORT`: The port number (default: `8080`)
- `GITHUB_APP_PRIVATE_KEY`: A GitHub App Private Key
- `WEBHOOK_SECRET`: A Webhook Secret
- `SECRET`: A YAML string for secrets
- `SECRET_FILE`: A secret file path
- `REQUEST_ID_HEADER`: A HTTP Header for request id. In case of Google Cloud Function `X-Cloud-Trace-Context` is used.

[See Secrets for `SECRET` and `SECRET_FILE`.](secret.md)
