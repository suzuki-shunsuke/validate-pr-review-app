# Configuration

Configuration consists of **secrets** and **non-secrets**.

## Secrets

- `webhook_secret`
- `github_app_private_key`

> [!WARNING]
> When using AWS Secrets Manager Web UI, multi-line values are not supported.
> You should convert the private key and webhook secret into JSON before storing.

## Non Secret Config

You can configure AWS Lambda Function by environment variable `CONFIG`.
`CONFIG` is a YAML string.

## Environment Variables

Either `CONFIG` or `CONFIG_FILE` is required.

- `CONFIG`: A YAML string for configuration
- `CONFIG_FILE`: A configuration file path

For HTTP Server:

- `PORT`: The port number (default: `8080`)
- `GITHUB_APP_PRIVATE_KEY`: A GitHub App Private Key
- `WEBHOOK_SECRET`: A Webhook Secret
- `SECRET`: A YAML string for secrets
- `SECRET_FILE`: A secret file path
- `REQUEST_ID_HEADER`: A HTTP Header for request id. In case of Google Cloud Function `X-Cloud-Trace-Context` is used.

```yaml
webhook_secret: 0123456789abcdefghijklmn
github_app_private_key: |
  -----BEGIN RSA PRIVATE KEY-----
  ...
```

## JSON Schema

[json-schema/config.json](../json-schema/config.json)

You can validate your config using JSON Schema and tools such as [ajv-cli](https://ajv.js.org/packages/ajv-cli.html).

```sh
ajv --spec=draft2020 -s json-schema/config.json -d config.yaml
```

### Input Complementation by YAML Language Server

[Please see the comment too.](https://github.com/szksh-lab/.github/issues/67#issuecomment-2564960491)

Version: `main`

```yaml
# yaml-language-server: $schema=https://raw.githubusercontent.com/suzuki-shunsuke/validate-pr-review-app/main/json-schema/config.json
```

Or pinning version:

```yaml
# yaml-language-server: $schema=https://raw.githubusercontent.com/suzuki-shunsuke/validate-pr-review-app/v0.0.1/json-schema/config.json
```

## Example Config

> [!WARNING]
> Please remove `[bot]` from each app name of `trusted_apps`
> :o: `dependabot`
> :x: `dependabot[bot]`

```yaml
# yaml-language-server: $schema=https://raw.githubusercontent.com/suzuki-shunsuke/validate-pr-review-app/main/json-schema/config.json
# Required
app_id: 0000 # GitHub App ID
installation_id: 00000000 # GitHub App Installation ID
aws:
  secret_id: request-pr-review-app # Secret ID in AWS Secrets Manager
  use_lambda_function_url: true # Optional. true when using Lambda Function URL. Default: false

# Optional
check_name: check-approval # Optional. Default: validate-review
log_level: info # debug, info, warn, error. Default: info
trust:
  trusted_apps:
    - renovate
    - dependabot
  untrusted_machine_users:
    - "*-bot"
    - octocat
  trusted_machine_users:
    - suzuki-shunsuke-bot
repositories:
  # Repository specific config
  # Override the root config
  # Only the first element matching the repository is used
  # If no element matches, the root config is used
  - repositories:
      # Glob pattern matching repository full names
      - suzuki-shunsuke/*
    trust:
      untrusted_machine_users:
        - "*-bot"
        - bot-*
```

## :bulb: Customize footer

You can customize the footer of this app's Checks tab.

The default is: [footer.md](../pkg/config/templates/footer.md)

For example, you can add the guide for developers:

```yaml
templates:
  footer: |
    ---

    For more details, see the [document](https://example.com).
    If you have any questions, please contact the security team.
```

This template is rendered with [Go's html/template](https://pkg.go.dev/html/template).

## Allow Unsigned Commits

> [!WARNING]
> This setting isn't recommended in terms of security.

In the real world, sometimes it's hard to enforce signed commits.
So you may want to allow unsigned commits from certain authors.

- `allow_unsigned_commits`: If this is true, all unsigned commits don't require two approvals. By default, this is false.
- `unsigned_commit_apps`: If this is set, commits from these apps don't require two approvals. By default, this is empty.
  - If `allow_unsigned_commits` is true, please don't set this setting. Otherwise, the app fails.
  - Glob is unavailable
  - `[bot]` should be removed from login
- `unsigned_commit_machine_users`: If this is set, commits from these machine users don't require two approvals. By default, this is empty.
  - If `allow_unsigned_commits` is true, please don't set this setting. Otherwise, the app fails.
  - Glob is unavailable

```yaml
insecure:
  allow_unsigned_commits: true
repositories:
  - repositories:
      - suzuki-shunsuke/*
    insecure: # The repository config overrides the root config.
      unsigned_commit_machine_users:
        - "foo-bot"
      unsigned_commit_apps:
        - "foo-bot"
```
