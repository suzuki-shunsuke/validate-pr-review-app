---
name: validate-pr-review-app
description: |
  Use this skill when working with validate-pr-review-app, a self-hosted GitHub App that validates Pull Request reviews to ensure PRs cannot be merged without proper approvals.
  Use this skill when the user wants to:
  - Update the configuration of validate-pr-review-app (e.g. make an app trusted, add untrusted_machine_users, change approval requirements)
  - Understand why a PR requires two approvals or how to avoid the two-approval requirement
  - Troubleshoot PR merge failures related to approval validation
  - Deploy or run the app (AWS Lambda, HTTP Server), register its GitHub App, or verify its release assets and container images
  Even if the user doesn't mention "validate-pr-review-app" by name — if they ask about PR approval requirements, trusted apps, or merge validation, this skill applies.
---

validate-pr-review-app is a self-hosted GitHub App.
It receives `pull_request_review` and `pull_request` webhooks, fetches the PR's reviews and commits
via the GitHub API, validates the reviews, and reports the result as a GitHub Check.
A PR needs at least 1 approval, and 2 approvals when a committer approves it or when the PR contains
unsigned commits, commits not linked to a GitHub user, or commits from untrusted Machine Users or
GitHub Apps.

Don't read every reference file. Read only the one that matches the task.

## Gotchas

- `trust.trusted_apps` doesn't support glob or regular expressions. A name containing `.`, `*`, `?`, `^`, `+`, or `$` makes the app fail at startup. `trust.untrusted_machine_users` does support glob, and a leading `!` marks a machine user as trusted (like `.gitignore`).
- Trust configuration is replaced, not merged, between the repository, global, and default levels. Setting `trust.trusted_apps` without listing `renovate` and `dependabot` makes them untrusted, because it replaces the default. `trusted_apps` and `untrusted_machine_users` are resolved independently of each other.
- Only the first `repositories` element matching the repository is used. The remaining matching elements are ignored.
- `insecure.allow_unsigned_commits: true` combined with `unsigned_commit_apps` or `unsigned_commit_machine_users` makes the app fail at startup. Use one or the other.
- A repository with `ignored: true` gets no check at all, so a required status check on it never becomes successful.

## Understanding and troubleshooting

- [Validation Rules](references/validation-rules.md) — what the app validates and when a PR passes or fails.
- [Why are 2 approvals required for a pull request?](references/why-2-approvals-required.md) — to explain or troubleshoot a PR that needs two approvals, including a check that stays `pending`.
- [How To Avoid 2 Approvals](references/how-to-avoid-2-approvals.md) — to reduce a PR's approval requirement, including using CSM Actions for automatic code fixes and approvals.
- [How It Works](references/how-it-works.md) — to understand the overall mechanism and the webhook flow.
- [Allow Empty Commits and Trivial Merge Commits](references/allow-empty-commit-and-trivial-merge-commit.md) — when an empty commit or an "update branch" merge commit unexpectedly does or doesn't require a second approval.
- [Handling Pull Request Events](references/handle-pull-request-event.md) — when no check is created after pushing a commit to an already-approved PR.
- [Merge Queue Support](references/merge-queue.md) — when the repository uses a merge queue.

## Configuration

- [Configuration](references/config.md) — to change configuration (`trusted_apps`, `untrusted_machine_users`, trust priority between repository/global/default config, `check_name`, `log_level`, `insecure` unsigned-commit rules) or to validate a config file with its JSON Schema.
- [Environment Variables](references/env.md) — to configure how the app process runs (`CONFIG`, `CONFIG_FILE`, `PORT`, `SECRET_FILE`, etc.).
- [Secrets](references/secret.md) — to set up `webhook_secret` and `github_app_private_key`, and to choose a secret source (env var, file, AWS Secrets Manager, Google Secret Manager).
- [Customize footer](references/customize-footer.md) — to customize the footer shown on the Checks tab.

## Setup and operation

- [GitHub App Settings](references/github-app.md) — to register the GitHub App with the right permissions and event subscriptions.
- [Getting Started - HTTP Server](references/getting-started-http.md) — to run the app locally as an HTTP server and receive webhooks via smee.io. Also lists the HTTP endpoints.
- [Getting Started - AWS Lambda](references/getting-started-lambda.md) — to deploy the app to AWS Lambda with Terraform, using a Function URL or Amazon API Gateway.
- [Logging, Monitoring, and Security](references/production.md) — to run the app in production, and to read or alert on its JSON logs.
- [Verify Release Assets](references/verify-asset.md) — to verify downloaded release assets with GitHub CLI, slsa-verifier, or Cosign.
- [Verify Container Images](references/verify-image.md) — to verify container images with Cosign.
