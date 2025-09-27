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
2. GitHub sends Webhook to the App when pull requests are reviewed or pull requests are added to merge queue.
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

Deploying the app to AWS Lambda using Terraform.

Requirements:

- AWS Account
- Terraform
- Git
- bash
- GitHub CLI
- GitHub Repository where the app is installed.
  - A pull request created by other than you is necessary

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

4. [Registering a GitHub App](https://docs.github.com/en/apps/creating-github-apps/registering-a-github-app/registering-a-github-app)

- Deactivate Webhook for now. We'll enable this after deploying the AWS Lambda Function.
- Permissions:
  - Checks: Read and write
  - Contents: Read-only
  - Pull requests: Read-only
- `Where can this GitHub App be installed?` > `Only on this account`

After registering the app, you can get the app id from the setting page.
Please add it to config.yaml.

```yaml
app_id: 0123456
```

5. [Generate a Private Key of the GitHub App.](https://docs.github.com/en/apps/creating-github-apps/authenticating-with-a-github-app/managing-private-keys-for-github-apps)

6. Add the private key to [secret.yaml](terraform/aws/secret.yaml.tmpl) and remove the downloaded private key file.

secret.yaml

```yaml
github_app_private_key: |
  -----BEGIN RSA PRIVATE KEY-----
  ...
```

7. [Install the app to your repository](https://docs.github.com/en/apps/using-github-apps/installing-your-own-github-app)

Please install the app to your repository.
[If you don't have any repository for this, please create a repository.](https://github.com/new)

After installing the app, you can get the installation id from URL.
Please add it to config.yaml.

```yaml
installation_id: 01234567
```

8. [Create a secret token](https://docs.github.com/en/webhooks/using-webhooks/validating-webhook-deliveries) and add it to secret.yaml

secret.yaml

```yaml
webhook_secret: 0123456789abcdefghijklmn
github_app_private_key: |
  -----BEGIN RSA PRIVATE KEY-----
  ...
```

9. Deploy the app by Terraform

(Optional) If you want to change input variables, please check [variables.tf](terraform/aws/variables.tf) and create a file `terraform.tfvars`.

e.g.

```
region = "ap-northeast-1" # Default: us-east-1
```

Then running Terraform commands.

```sh
terraform init
terraform validate
terraform plan
terraform apply
```

10. [Configure Webhook](https://docs.github.com/en/apps/creating-github-apps/registering-a-github-app/using-webhooks-with-github-apps)

You can check the webhook URL by `terraform state show`.

```console
$ terraform state show aws_lambda_function_url.main
# aws_lambda_function_url.main:
resource "aws_lambda_function_url" "main" {
    # ...
    function_url       = "https://abcdefghijklmnopqrstuvwxyz012345.lambda-url.us-east-1.on.aws/"
    # ...
}
```

- Set the Lambda Function URL to the webhook URL
- Set the secret token to the webhook secret

Please don't forget to click `Save changes`.

11. Subscribe Events for GitHub App

Checks the following events.

- Pull request review

And click `Save changes`.

If the button `Save changes` is disabled and you can't clike `Save changes`, please try to change any permission and revert the change.

The setup was done.
You can create reviews and try the app.
Please prepare a pull request that you haven't pushed any commit.

12. Approve a pull request

Then the check passes.

If the app doesn't work, please check the AWS CloudWatch Log to check if the request reached to the AWS Lambda.

13. Dismiss the review.

Then the check fails.

14. Clean up

You can destroy resources by `terraform destroy`.

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
    - `checks:write` To create Check Runs by GitHub API
    - `pull_requests:read`: To retrieve pull request reviews by GitHub API
    - `contents:read`: To retrieve pull request commits by GitHub API
  - Private Key (keep safe).
- Deploy Validate PR Review App.
- Store Webhook Secret & GitHub App Private Key in **AWS Secrets Manager**.
- Enable Webhooks in your GitHub App:
  - Set the Webhook Secret.
  - Point the Webhook URL to your Lambda.
- Install the App in repositories to validate PRs.
- [Add the check to Required Checks by GitHub Branch Rulesets](https://docs.github.com/en/repositories/configuring-branches-and-merges-in-your-repository/managing-rulesets/about-rulesets)

## Merge Queue Support

This app supports [Merge Queue](https://docs.github.com/en/repositories/configuring-branches-and-merges-in-your-repository/configuring-pull-request-merges/managing-a-merge-queue).
Additional settings aren't necessary.

## Configuration

Configuration consists of **secrets** and **non-secrets**.

### Secrets

- `webhook_secret`
- `github_app_private_key`

> [!WARNING]
> When using AWS Secrets Manager Web UI, multi-line values are not supported.
> You should convert the private key and webhook secret into JSON before storing.

### Non Secret Config

You can configure AWS Lambda Function by environment variable `CONFIG`.
`CONFIG` is a YAML string.

#### JSON Schema

[json-schema/config.json](json-schema/config.json)

You can validate your config using JSON Schema and tools such as [ajv-cli](https://ajv.js.org/packages/ajv-cli.html).

```sh
ajv --spec=draft2020 -s json-schema/config.json -d config.yaml
```

##### Input Complementation by YAML Language Server

[Please see the comment too.](https://github.com/szksh-lab/.github/issues/67#issuecomment-2564960491)

Version: `main`

```yaml
# yaml-language-server: $schema=https://raw.githubusercontent.com/suzuki-shunsuke/validate-pr-review-app/main/json-schema/config.json
```

Or pinning version:

```yaml
# yaml-language-server: $schema=https://raw.githubusercontent.com/suzuki-shunsuke/validate-pr-review-app/v0.0.1/json-schema/config.json
```

### Example Config

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

### :bulb: Customize footer

You can customize the footer of this app's Checks tab.

The default is:

```markdown
---

[This check is created by Validate PR Review App (version: {{if .Version}}{{.Version}}{{else}}unknown{{end}})](https://github.com/suzuki-shunsuke/validate-pr-review-app).
```

You can add the guide for developers:

e.g.

```yaml
templates:
  footer: |
    ---

    For more details, see the [document](https://example.com).
    If you have any questions, please contact the security team.
```

This template is rendered with [Go's html/template](https://pkg.go.dev/html/template).

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

## Using CSM Actions For Secure Automatic Code Fixes and Approvals

By using the **Validate PR Review App**, you can prevent commits and approvals made by untrusted Apps or Machine Users.
However, requiring two approvals every time CI automatically fixes code can hurt developer productivity.

[**CSM Actions**](https://github.com/csm-actions/docs) solves this problem.
CSM Actions is a collection of GitHub Actions that securely handle code modifications and approvals through a **Client/Server Model**.
With this model, sensitive credentials such as a GitHub App’s Private Key or a Machine User’s Personal Access Token never need to be passed to the client side (regular GitHub Actions workflows). Instead, they are securely managed on the server side (a centrally managed GitHub repository and workflow).

Here are some available Actions:

- [**Securefix Action**](https://github.com/csm-actions/securefix-action): Securely create commits and pull requests.
- [**Approve PR Action**](https://github.com/csm-actions/approve-pr-action): Securely approve PRs using a Machine User.
- [**Update Branch Action**](https://github.com/csm-actions/update-branch-action): Securely update PR branches.
  - If a reviewer updates a branch from the GitHub Web UI, another reviewer’s approval is required to prevent self-approval. With Update Branch Action, the branch is updated securely using a GitHub App.

By registering the Apps or Machine Users used with CSM Actions in `trusted_apps` or `untrusted_machine_users`, you can achieve automatic code fixes and auto-merge without additional PR reviews.

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
