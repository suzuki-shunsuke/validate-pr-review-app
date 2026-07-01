---
name: validate-pr-review-app-configuration
description: |
  Configure validate-pr-review-app. Covers the config file and JSON Schema, environment
  variables (CONFIG, CONFIG_FILE, PORT, SECRET_FILE, etc.), secrets (webhook_secret,
  github_app_private_key), the trust model (trusted_apps, untrusted_machine_users), per-repository
  config overrides, footer template customization, and allowing unsigned commits.
  Use this skill when the user wants to:
  - Change the app's configuration — make an app trusted, add untrusted_machine_users, set the
    check name or log level, or add a per-repository override.
  - Set up secrets or environment variables for the app.
  - Customize the Checks footer or allow unsigned commits from specific authors.
  For how the approval rules themselves behave, use the validation skill instead.
---

Read [reference.md](reference.md) to configure validate-pr-review-app — the trust model
(trusted vs. untrusted apps and machine users), secrets, environment variables, JSON Schema,
an example config with per-repository overrides, footer customization, and the unsigned-commit
allowlist settings.
