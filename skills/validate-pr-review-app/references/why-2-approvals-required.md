# Why are 2 approvals required for a pull request?

Please check the check api.

```sh
repo=xxx
pr_number=xxx
check_name=validate-review

gh pr view "$pr_number" -R "$repo" --json headRefOid --jq '.headRefOid' |
  xargs -I{} gh api "repos/${repo}/commits/{}/check-runs" --jq '.check_runs[] | select(.name == "'"$check_name"'") | .id' |
  xargs -I{} gh api "repos/${repo}/check-runs/{}" --jq '.output.summary'
```

Please see [How To Avoid 2 Approvals](how-to-avoid-2-approvals.md) and [Validation Rules](validation-rules.md) too.

If the check is `pending`, please check the app works properly.

1. Check the app works fine
1. Check if the app is installed into the repository
1. Trigger the app by approving the pull request. GitHub allows the same review to create multiple approvals.
    1. Maybe the app couldn't update check for some reasons. In that case, retrying would solve the problem
