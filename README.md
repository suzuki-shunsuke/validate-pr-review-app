# Validate PR Review App

[![Ask DeepWiki](https://deepwiki.com/badge.svg)](https://deepwiki.com/suzuki-shunsuke/validate-pr-review-app) [Documentation](#documentation) | [Agent Skills](#agent-skills)

Validate PR Review App is a self-hosted GitHub App that validates Pull Request reviews.
It helps organizations improve governance and security by ensuring PRs cannot be merged without proper approvals while keeping developer experience.

![image](https://github.com/user-attachments/assets/68e6fd3f-b36a-4d62-a46c-76bbeaf1ebdb)

![image](https://github.com/user-attachments/assets/aa460dc1-375c-46ad-ad05-24cdea7f1c4d)

## Features

- Security and Governance
  - Enforce Pull Request reviews
  - Centralized configuration: Manage settings in one place via the GitHub App, keeping governance and security strong with minimal overhead.
- Good Developer Experience
  - Runs quickly and provides clear error feedback through the Checks API, so developers immediately understand why validation failed.

### Validation Rules

- At least **1 approval** required.
- If the committer approves → **2 approvals required**.
  - [As of v0.3.2, empty commits and trivial merge commits don't require 2 approvals](skills/validate-pr-review-app/references/allow-empty-commit-and-trivial-merge-commit.md)
- If the PR contains [unsigned commits](https://docs.github.com/en/authentication/managing-commit-signature-verification/signing-commits) or [commits not linked to a GitHub user](https://docs.github.com/en/pull-requests/committing-changes-to-your-project/troubleshooting-commits/why-are-my-commits-linked-to-the-wrong-user) → **2 approvals required**.
- Approvals from untrusted Machine Users or GitHub Apps are ignored.
- If the PR contains commits from untrusted Machine Users or GitHub Apps → **2 approvals required**.

[See Validation Rules for details.](skills/validate-pr-review-app/references/validation-rules.md)

## How It Works

Install the GitHub App in your repositories and enable Webhook.
GitHub sends a Webhook to the App when pull requests are reviewed or added to a merge queue.
The App validates the Webhook, ignores irrelevant events, fetches the PR's reviews and commits via the GitHub API, validates the reviews, and updates the Check via the Checks API.

[See How It Works for the detailed flow.](skills/validate-pr-review-app/references/how-it-works.md)

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

- AWS Lambda
  - Function URL
  - Amazon API Gateway
- HTTP Server

## Getting Started

- [Run HTTP Server in your localhost](skills/validate-pr-review-app/references/getting-started-http.md)
- [AWS Lambda](skills/validate-pr-review-app/references/getting-started-lambda.md)

## Merge Queue Support

This app supports [Merge Queue](https://docs.github.com/en/repositories/configuring-branches-and-merges-in-your-repository/configuring-pull-request-merges/managing-a-merge-queue).
[Additional settings aren't necessary.](skills/validate-pr-review-app/references/merge-queue.md)

## Trusted vs. Untrusted Users and GitHub Apps

- **Trusted Apps & Users**: properly managed, cannot be abused.
- **Untrusted Apps & Users**: potentially exploitable (e.g., private keys exposed).

By default, `renovate` and `dependabot` are trusted Apps, and Machine Users are trusted unless configured otherwise.

```yaml
trust:
  trusted_apps:
    - renovate
    - dependabot
  untrusted_machine_users:
    - "*-bot"
    - "!my-safe-bot" # exclude from the pattern above
```

[See Configuration for how trust is resolved between repository, global, and default configuration.](skills/validate-pr-review-app/references/config.md#trustness)

## Using CSM Actions For Secure Automatic Code Fixes and Approvals

By using the **Validate PR Review App**, you can prevent commits and approvals made by untrusted Apps or Machine Users.
However, requiring two approvals every time CI automatically fixes code can hurt developer productivity.

[**CSM Actions**](https://github.com/csm-actions/docs) solves this problem.
CSM Actions is a collection of GitHub Actions that securely handle code modifications and approvals through a **Client/Server Model**.
With this model, sensitive credentials such as a GitHub App’s Private Key or a Machine User’s Personal Access Token never need to be passed to the client side (regular GitHub Actions workflows). Instead, they are securely managed on the server side (a centrally managed GitHub repository and workflow).

[See How To Avoid 2 Approvals for the available Actions and how to combine them with `trusted_apps`.](skills/validate-pr-review-app/references/how-to-avoid-2-approvals.md)

## Agent Skills

[About Agent Skills, please see the official document.](https://agentskills.io/home)

This repository provides an Agent Skill for validate-pr-review-app: [skills/validate-pr-review-app/SKILL.md](skills/validate-pr-review-app/SKILL.md)

Install the skill using [gh skill install](https://cli.github.com/manual/gh_skill_install):

```sh
gh skill install suzuki-shunsuke/validate-pr-review-app --all
```

## Documentation

The detailed documentation is split by topic and lives in [skills/validate-pr-review-app/references](skills/validate-pr-review-app/references).
These files are the single source of truth, shared between this README and the Agent Skill, so there's no duplicated maintenance.

Understanding and troubleshooting:

- [Validation Rules](skills/validate-pr-review-app/references/validation-rules.md) - what the app validates and when a PR passes or fails.
- [Why are 2 approvals required for a pull request?](skills/validate-pr-review-app/references/why-2-approvals-required.md) - diagnose a PR that needs two approvals.
- [How To Avoid 2 Approvals](skills/validate-pr-review-app/references/how-to-avoid-2-approvals.md) - reduce a PR's approval requirement, including CSM Actions.
- [How It Works](skills/validate-pr-review-app/references/how-it-works.md) - the webhook flow from event to Check.
- [Allow Empty Commits and Trivial Merge Commits](skills/validate-pr-review-app/references/allow-empty-commit-and-trivial-merge-commit.md) - why these commits don't require a second approval, and how they're detected.
- [Handling Pull Request Events](skills/validate-pr-review-app/references/handle-pull-request-event.md) - how checks are created when commits are pushed to an approved PR.
- [Merge Queue Support](skills/validate-pr-review-app/references/merge-queue.md) - using the app with a merge queue.

Configuration:

- [Configuration](skills/validate-pr-review-app/references/config.md) - JSON Schema, example config, trust resolution, and unsigned commit settings.
- [Environment Variables](skills/validate-pr-review-app/references/env.md) - how the app process is configured.
- [Secrets](skills/validate-pr-review-app/references/secret.md) - secret sources and the required secrets.
- [Customize footer](skills/validate-pr-review-app/references/customize-footer.md) - customize the footer shown on the Checks tab.

Setup and operation:

- [GitHub App Settings](skills/validate-pr-review-app/references/github-app.md) - permissions and events to subscribe.
- [Getting Started - HTTP Server](skills/validate-pr-review-app/references/getting-started-http.md) - run the app locally and receive webhooks via smee.io.
- [Getting Started - AWS Lambda](skills/validate-pr-review-app/references/getting-started-lambda.md) - deploy to AWS Lambda with Terraform.
- [Logging, Monitoring, and Security](skills/validate-pr-review-app/references/production.md) - running the app in production.
- [Verify Release Assets](skills/validate-pr-review-app/references/verify-asset.md) - verify downloaded release assets.
- [Verify Container Images](skills/validate-pr-review-app/references/verify-image.md) - verify container images with Cosign.

## License

[MIT](LICENSE)
