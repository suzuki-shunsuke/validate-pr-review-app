# Validate PR Review App

Validate PR Review App is a self-hosted GitHub App that validates Pull Request reviews.
It helps organizations improve governance and security by ensuring PRs cannot be merged without proper approvals while keeping developer experience.

![image](https://github.com/user-attachments/assets/68e6fd3f-b36a-4d62-a46c-76bbeaf1ebdb)

![image](https://github.com/user-attachments/assets/aa460dc1-375c-46ad-ad05-24cdea7f1c4d)

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

### Validation Rules

- At least **1 approval** required.
- If the committer approves → **2 approvals required**.
- If the PR contains [unsigned commits](https://docs.github.com/en/authentication/managing-commit-signature-verification/signing-commits) or [commits not linked to a GitHub user](https://docs.github.com/en/pull-requests/committing-changes-to-your-project/troubleshooting-commits/why-are-my-commits-linked-to-the-wrong-user) → **2 approvals required**.
- Approvals from untrusted Machine Users or GitHub Apps are ignored.
- If the PR contains commits from untrusted Machine Users or GitHub Apps → **2 approvals required**.

## How It Works

1. Install the GitHub App in your repositories and [enable Webhook](https://docs.github.com/en/apps/creating-github-apps/registering-a-github-app/using-webhooks-with-github-apps).
2. GitHub sends Webhook to the App when pull requests are reviewed.
3. The App validates if the Webhook is valid.
4. The App filters irrevant events like review comments.
5. The App fetches PR reviews and commits using the GitHub API.
6. The App validates reviews.
7. The App updates the Check via the Checks API.

```mermaid
sequenceDiagram
    participant GitHub
    participant ValidatePRReviewApp as Validate PR Review App

    GitHub ->> ValidatePRReviewApp: Send Pull Request Review Webhook
    ValidatePRReviewApp ->> ValidatePRReviewApp: Validate Webhook
    ValidatePRReviewApp ->> ValidatePRReviewApp: Ignore irrelevant events
    ValidatePRReviewApp ->> GitHub: Fetch PR reviews and commits (GitHub API)
    GitHub -->> ValidatePRReviewApp: Reviews & commits data
    ValidatePRReviewApp ->> ValidatePRReviewApp: Validate Reviews
    ValidatePRReviewApp ->> GitHub: Update Check (Checks API)
```

## Why?

This project is the successor to the following our OSS Projects:

1. [deny-self-approve](https://github.com/suzuki-shunsuke/deny-self-approve) (CLI)
2. [validate-pr-review-action](https://github.com/suzuki-shunsuke/validate-pr-review-action) (GitHub Action)

When developing as a team, it's common to require that pull requests be reviewed by someone other than the author.
Code reviews help improve code quality, facilitate knowledge sharing among team members, and prevent any single person from making unauthorized changes without approval.

First, you should enable the following branch ruleset on the default branch.

- `Require a pull request before merging`
  - `Require review from Code Owners`
  - `Require approval of the most recent reviewable push`
- `Require status checks to pass`

This rules require pull request reviews, but there are still several ways to improperly merge a pull request without a valid review:

1. Abusing a machine user with `CODEOWNER` privileges to approve the PR.
2. Adding commits to someone else’s PR and approving it yourself.
3. Using a machine user or bot to add commits to someone else’s PR, then approving it yourself.

[validate-pr-review-action](https://github.com/suzuki-shunsuke/validate-pr-review-action) validates pull request reviews via `pull_request_review` or `merge_group` events.
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

**Validate PR Review App** solves these issues by working as a GitHub App, receiving Webhooks, and updating Checks directly.

## Supported Platforms

Now only AWS Lambda is supported.

## Getting Started

In the getting started, we will deploy the app to AWS Lambda using Terraform.

Requirements:

- AWS Account
- Terraform
- Git
- bash
- GitHub CLI
- GitHub Repository where the app is installed.
  - A pull request created by other than you is necessary

Of course, we use these tools in the getting started, but these tools aren't mandatory for the app.
You can deploy the app without them.

After the getting started, you can destroy all resources using Terraform.

1. Checkout the repository

```sh
git clone https://github.com/suzuki-shunsuke/validate-pr-review-app
```

2. Move to [terraform/aws](terraform/aws).

```sh
cd validate-pr-review-app/terraform/aws
```

3. Run `bash init.sh`

```sh
bash init.sh
```

4. Create a GitHub App

Permissions:

- `checks:write`
- `pull_requests:read`
- `contents:read`

Create a Private Key.

5. Edit [config.yaml](terraform/aws/config.yaml.tmpl) and [secret.yaml](terraform/aws/secret.yaml.tmpl).

At least, you need to add the private key to secret.yaml.

```yaml
github_app_private_key: |
  -----BEGIN RSA PRIVATE KEY-----
  ...
```

6. Deploy the app by Terraform

```sh
terraform init
terraform validate
terraform plan
terraform apply
```

7. Install the app to a repository
8. [Configure Webhook](https://docs.github.com/en/apps/creating-github-apps/registering-a-github-app/using-webhooks-with-github-apps)

- Set the Lambda Function URL to the webhook URL
- Set the secret token `dummy-secret` to the webhook secret

The setup was done.
You can create reviews and try the app.

9. Approve a pull request

Then the check passes.

If the app doesn't work, please check the AWS CloudWatch Log to check if the request reached to the AWS Lambda.

10. Dismiss the review.

Then the check fails.

11. Clean up

You can destory resources by `terraform destroy`.

```sh
terraform destroy
```

- Uninstall the GitHub App from the repository
- Delete the GitHub App

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
- Deploy Validate PR Review App.
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

> [!WARNING]
> Please remove `[bot]` from each app name of `trusted_apps`
> :o: `dependabot`
> :x: `dependabot[bot]`

```yaml
app_id: 0000 # GitHub App ID
installation_id: 00000000 # GitHub App Installation ID
aws:
  secret_id: request-pr-review-app # Secret ID in AWS Secrets Manager
  use_lambda_function_url: true # true when using Lambda Function URL
check_name: check-approval # Optional. Default: validate-review
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

## Logging, Monitoring, Security, etc

Please see the following documents:

- AWS Lambda
  - [Sending Lambda function logs to CloudWatch Logs](https://docs.aws.amazon.com/lambda/latest/dg/monitoring-cloudwatchlogs.html)
  - [Monitoring, debugging, and troubleshooting Lambda functions](https://docs.aws.amazon.com/lambda/latest/dg/lambda-monitoring.html)
- GitHub App
  - [Validating webhook deliveries](https://docs.github.com/en/webhooks/using-webhooks/validating-webhook-deliveries)

The log format of Validate PR Review App is the JSON format.
The log has the log level like `INFO`, `WARN`, and `ERROR`, so you can send alerts based on the log level.

<details>
<summary>Example Log</summary>

```json
{
    "time": "2025-09-25T19:49:28.295812986Z",
    "level": "INFO",
    "msg": "Fetched a pull request",
    "version": "",
    "pull_request": {
        "sha": "e21cc0e643655273c71f1d14e3f42ee14c2c6721",
        "approvers": {
            "suzuki-shunsuke": {}
        },
        "commits": [
            {
                "oid": "25031c1e9c20594e0dc4569e193ad3f45de0ade1",
                "committer": {
                    "login": "renovate[bot]",
                    "is_app": true
                },
                "signature": {
                    "isValid": true,
                    "state": "VALID"
                }
            }
        ]
    }
}
```

</details>

## License

[MIT](LICENSE)
