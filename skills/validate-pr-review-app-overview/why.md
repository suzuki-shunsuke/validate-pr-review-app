## Why?

This project is the successor to [deny-self-approve](https://github.com/suzuki-shunsuke/deny-self-approve) (CLI) and [validate-pr-review-action](https://github.com/suzuki-shunsuke/validate-pr-review-action) (GitHub Action).
Even with branch rulesets that require reviews, PRs can still be merged improperly — for example by abusing a machine user with `CODEOWNER` privileges, or by adding commits to someone else's PR and approving it yourself.
GitHub Actions-based validation doesn't scale well for larger organizations (setup cost, easy to bypass, slower, poor error visibility), so this app solves these issues by working as a GitHub App, receiving Webhooks, and updating Checks directly.
