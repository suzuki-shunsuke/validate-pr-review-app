# Allow Empty Commits and Trivial Merge Commits

As of v0.3.2, validate-pr-review-app doesn't require 2 approvals for empty commits and trivial merge commits by approvers.

## What are trivial merge commits?

When a PR's base branch can be merged into the feature branch without any conflict resolution, the resulting merge commit is referred to here as a "trivial merge commit".
The term "trivial merge" is inspired by the following documentation, but the definition of a trivial merge commit in validate-pr-review-app is not strictly the same.

https://git-scm.com/docs/trivial-merge

## Why?

To prevent PRs from being merged without review by others, validate-pr-review-app requires approval from someone other than the approver if the approver has commits in the PR.
However, empty commits and trivial merge commits do not contain changes that need to be reviewed.
Empty commits obviously have no changes, and the changes in trivial merge commits are already included in the base branch, so merging them into the base branch causes no issues.
Therefore, since v0.3.2, even if these commits are present, a second approval is no longer required.

## How are trivial merge commits detected?

Here is how validate-pr-review-app determines whether a commit is a trivial merge commit.
Since validate-pr-review-app does not depend on Git, it retrieves commit information via the GitHub API.
Note that this results in GitHub API calls that consume rate limits.

First, when fetching PR commits and reviews via the GraphQL API, the number of changed files (`changedFilesIfAvailable`) and parent SHAs are also retrieved.
If `changedFilesIfAvailable` is 0, the commit is an empty commit.
Next, the number of parents is checked to ensure there are exactly 2.
A trivial merge commit always has exactly two parents: the first parent (the previous commit on the feature branch) and a commit from the PR's base branch.
If there are 3 or more parents, branches other than the base branch have been merged, which means unreviewed changes may be included and review is required.
The diffs between each parent SHA and the commit are retrieved using the [Compare two commits API](https://docs.github.com/en/rest/commits/commits), and the changed files are checked for overlaps.
If there are overlapping files, conflict resolution has likely occurred, and review is needed to verify the resolution is correct.
Strictly speaking, different parts of the same file may have been modified without causing a conflict, but checking whether code changes overlap at the line level via the API would be overly complex. So as a safety measure, if the same file appears in both diffs, review is required.
Due to API limitations, the maximum number of files that can be retrieved is 300. If 300 or more files are changed, it is treated conservatively as having overlaps, and review is required.
If a file other than the merged branch's files is modified when generating the merge commit, that file's diff appears in both the diff against the base branch and the diff against the first parent, resulting in an overlap.
Additionally, the non-first parent is compared with the HEAD of the PR's base branch using the [Compare two commits API](https://docs.github.com/en/rest/commits/commits) to check whether the parent is an ancestor of the HEAD.
If `behindBy` is 0, it is an ancestor.
If the parent is not a commit from the base branch, it means another feature branch was merged, and review is required.
