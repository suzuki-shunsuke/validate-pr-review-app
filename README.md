# Require PR Review App

**Require PR Review App** is a self-hosted GitHub App that enforces Pull Request reviews.
It helps organizations improve governance and security by ensuring PRs cannot be merged without proper approvals while keeping developer experience.

## :warning: Status

This project is still in alpha.
Probably this doesn't work.
Please don't use this yet.

## Features

- Security and Governance
  - Enforce Pull Request reviews
  - Centralized configuration: Manage settings in one place via the GitHub App, keeping governance and security strong with minimal overhead.
- Good Developer Experience
  - Runs quickly and provides clear error feedback through the Checks API, so developers immediately understand why validation failed.

## Supported Platforms

Now only AWS Lambda is supported.

### Validation Rules

- At least **1 approval** required.
- If the committer approves → **2 approvals required**.
- If the PR contains unsigned commits or commits not linked to a GitHub user → **2 approvals required**.
- Approvals from untrusted Machine Users or GitHub Apps are ignored.
- If the PR contains commits from untrusted Machine Users or GitHub Apps → **2 approvals required**.

## Why?

This project is the successor to:

1. [deny-self-approve](https://zenn.dev/shunsuke_suzuki/articles/deny-self-approve) (CLI)
2. [validate-pr-review-action](https://zenn.dev/shunsuke_suzuki/articles/validate-pr-review-action) (GitHub Action)

While GitHub Actions-based validation works for small projects, it doesn’t scale well for larger organizations due to:

- **Setup & management cost**
  - Workflows must be added and maintained in every repository.
  - Required Workflows don’t support the `pull_request_review` event.
- **Security & governance**
  - Easy to bypass by editing workflows.
  - Hard to centrally manage trusted apps or settings.
- **Developer experience**
  - Slower execution compared to FaaS (serverless).
  - Workflows trigger unnecessarily (e.g., on review comments).
  - Poor error visibility (logs instead of clear feedback).

**Require PR Review App** solves these issues by working as a GitHub App, receiving Webhooks, and updating Checks directly.

## How It Works

1. Install the App in your repository.
2. GitHub sends **Pull Request Review Webhook events** to the App.
3. The App validates the Webhook (secret verification).
4. The App filters events (ignores irrelevant ones like review comments).
5. The App fetches PR reviews and commits using the GitHub API.
6. The App validates them and updates the Check via the Checks API.

## Setup

> [!WARNING]
> We will release pre-built binaries for AWS Lambda to GitHub Releases.
> Now you need to build them yourself.

- [Generate a Webhook Secret Token](https://docs.github.com/en/webhooks/using-webhooks/validating-webhook-deliveries).
- Create a dedicated GitHub App with:
  - Permissions:
    - `checks:write`
    - `pull_requests:read`
    - `contents:read`
  - Private Key (keep safe).
- Deploy Require PR Review App.
- Store Webhook Secret & GitHub App Private Key in **AWS Secrets Manager**.
- Enable Webhooks in your GitHub App:
  - Set the Webhook Secret.
  - Point the Webhook URL to your Lambda.
- Install the App in repositories to validate PRs.

## Configuration

Configuration consists of **secrets** and **non-secrets**.

### Secrets

- `webhook_secret`
- `github_app_private_key`

> [!WARNING]
> When using AWS Secrets Manager Web UI, multi-line values are not supported.
> You should convert the private key and webhook secret into JSON before storing.

### Example Config

```yaml
app_id: 0000 # GitHub App ID
installation_id: 00000000 # GitHub App Installation ID
aws:
  secret_id: request-pr-review-app # Secret ID in AWS Secrets Manager
  use_lambda_function_url: true # true when using Lambda Function URL
check_name: check-approval # Optional. Default: verify-approval
trusted_apps:
  - renovate
  - dependabot
untrusted_machine_users:
  - "*-bot"
  - octocat
trusted_machine_users:
  - suzuki-shunsuke-bot
```

## Trusted vs. Untrusted Users and GitHub Apps

- **Trusted Apps & Users**: properly managed, cannot be abused.
- **Untrusted Apps & Users**: potentially exploitable (e.g., private keys exposed).

By default:

- `renovate` and `dependabot` are trusted Apps.
- Machine Users are trusted unless configured otherwise.
  - This is because machine users can't be distinguished with normal users without configuration.

Example:

```yaml
trusted_apps:
  - renovate
  - dependabot
untrusted_machine_users:
  - "*-bot"
trusted_machine_users:
  - my-safe-bot
```

## Demo

- Approve a PR created by Renovate or Dependabot → Check passes.
- Dismiss a review → Check fails.
- Add a commit yourself and approve → Check fails (self-approval detected).
- Leave a review comment → No Check update.

## License

[MIT](LICENSE)
