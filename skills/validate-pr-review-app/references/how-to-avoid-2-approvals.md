# How To Avoid 2 Approvals

- [Sign Commits](https://docs.github.com/en/authentication/managing-commit-signature-verification/signing-commits)
- CI
  - Use [Securefix Action](https://github.com/csm-actions/securefix-action)
  - Use [Update Branch Action](https://github.com/csm-actions/update-branch-action)
    - If approvers use the `Update Branch` button, other approvals are required due to the validation rule `If a committer approves, 2 approvals are required.`
- Add GitHub Apps to `trusted_apps`
- [Push empty commits or trivial merge commits, which don't require 2 approvals](allow-empty-commit-and-trivial-merge-commit.md)
- [Not Recommended: Configure `insecure`](config.md#allow-unsigned-commits)

## Using CSM Actions For Secure Automatic Code Fixes and Approvals

By using the Validate PR Review App, you can prevent commits and approvals made by untrusted Apps or Machine Users.
However, requiring two approvals every time CI automatically fixes code can hurt developer productivity.

[CSM Actions](https://github.com/csm-actions/docs) solves this problem.
CSM Actions is a collection of GitHub Actions that securely handle code modifications and approvals through a Client/Server Model.
With this model, sensitive credentials such as a GitHub App's Private Key or a Machine User's Personal Access Token never need to be passed to the client side (regular GitHub Actions workflows). Instead, they are securely managed on the server side (a centrally managed GitHub repository and workflow).

Here are some available Actions:

- [Securefix Action](https://github.com/csm-actions/securefix-action): Securely create commits and pull requests.
- [Approve PR Action](https://github.com/csm-actions/approve-pr-action): Securely approve PRs using a Machine User.
- [Update Branch Action](https://github.com/csm-actions/update-branch-action): Securely update PR branches.
  - If a reviewer updates a branch from the GitHub Web UI, another reviewer's approval is required to prevent self-approval. With Update Branch Action, the branch is updated securely using a GitHub App.

By registering the Apps or Machine Users used with CSM Actions in `trusted_apps` or `untrusted_machine_users`, you can achieve automatic code fixes and auto-merge without additional PR reviews.

### Use Securefix Action

Securefix Action allows you to create commits and pull requests via GitHub Actions securely.
If you create commits using git commands or GitHub Actions like [stefanzweifel/git-auto-commit-action](https://github.com/stefanzweifel/git-auto-commit-action) and [peter-evans/create-pull-request](https://github.com/peter-evans/create-pull-request), you should replace them with Securefix Action.
And by adding the server GitHub App for Securefix Action to `trusted_apps`, you can avoid 2 approvals.
