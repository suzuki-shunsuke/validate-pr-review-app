# Validation Rules

- At least 1 approval is required.
- If a committer approves, 2 approvals are required.
  - [As of v0.3.2, empty commits and trivial merge commits don't require 2 approvals.](allow-empty-commit-and-trivial-merge-commit.md)
- If [some commits are unsigned](https://docs.github.com/en/authentication/managing-commit-signature-verification/signing-commits), 2 approvals are required.
- If [some commits are not linked to a GitHub user](https://docs.github.com/en/pull-requests/committing-changes-to-your-project/troubleshooting-commits/why-are-my-commits-linked-to-the-wrong-user), 2 approvals are required.
- Approvals from untrusted Machine Users or GitHub Apps are ignored.
- If some committers are untrusted, 2 approvals are required.

Whether a user or a GitHub App is trusted is decided by the `trust` configuration.
[See Configuration for details.](config.md#trustness)

## See Also

- [Why are 2 approvals required for a pull request?](why-2-approvals-required.md)
- [How To Avoid 2 Approvals](how-to-avoid-2-approvals.md)
- [Handling Pull Request Events](handle-pull-request-event.md)
