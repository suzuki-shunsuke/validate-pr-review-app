---
name: validate-pr-review-app-validation
description: |
  Explain how validate-pr-review-app validates Pull Request reviews — the approval rules, why a
  PR needs 1 or 2 approvals, how unsigned commits and untrusted apps/machine users affect the
  requirement, how Pull Request synchronize events are handled, empty/trivial merge commits,
  merge queue support, and CSM Actions for secure auto-fix/approval.
  Use this skill when the user wants to:
  - Understand or troubleshoot why a validate-pr-review-app check passed or failed.
  - Understand why a PR requires two approvals or how to avoid the two-approval requirement.
  - Understand how empty commits, trivial merge commits, or "update branch" are handled.
  Even if the user doesn't name validate-pr-review-app — if they ask about PR approval
  requirements or merge validation behavior, this skill applies. For changing settings such as
  trusted_apps, use the configuration skill instead.
---

Read the file that matches the task:

- Read [reference.md](reference.md) first — the validation rules (when 1 vs 2 approvals are required), how the app works end to end, merge queue support, and CSM Actions for automatic code fixes and approvals without extra reviews.
- Read [pull_request_events.md](pull_request_events.md) to understand how the app handles Pull Request `synchronize` events (v0.3.2+), walking back through parent commits, with a step-by-step worked example.
- Read [trivial_merge_commits.md](trivial_merge_commits.md) to understand what empty commits and trivial merge commits are, why they don't require a second approval, and how the app detects them via the GitHub API.
