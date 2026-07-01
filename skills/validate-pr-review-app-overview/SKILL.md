---
name: validate-pr-review-app-overview
description: |
  Understand what validate-pr-review-app is, what it does, and how it works overall.
  validate-pr-review-app is a self-hosted GitHub App that validates Pull Request reviews so PRs
  cannot be merged without proper approvals, improving governance and security while keeping
  developer experience. This is the entry-point skill: it covers the high-level overview, the
  end-to-end flow (webhook → fetch reviews/commits → validate → update Check), supported
  platforms, and why the app exists (the successor to deny-self-approve and
  validate-pr-review-action, and why a GitHub App scales better than GitHub Actions).
  Use this skill when the user asks what validate-pr-review-app is, what it does, how it works,
  or why to use it — and to find which other skill covers a specific topic (validation rules,
  configuration, GitHub App setup, operations, or verifying assets).
---

Read these files for the overview:

- [header.md](header.md) — what validate-pr-review-app is (one-paragraph summary).
- [how_it_works.md](how_it_works.md) — the end-to-end flow, with a sequence diagram.
- [why.md](why.md) — why the app exists and how it compares to deny-self-approve, validate-pr-review-action, and GitHub Actions.
- [features.md](features.md) — the security/governance and developer-experience features.
- [rules.md](rules.md) — a summary of the validation rules (see the validate-pr-review-app-validation skill for the full rules and behavior).
- [platforms.md](platforms.md) — supported platforms (AWS Lambda, HTTP Server).

For details on a specific topic, use the sibling skills: validate-pr-review-app-validation
(validation rules and behavior), validate-pr-review-app-configuration (settings and secrets),
validate-pr-review-app-github-app (registering the GitHub App), validate-pr-review-app-operations
(endpoints, logging, monitoring), and validate-pr-review-app-verify-assets (verifying assets).
