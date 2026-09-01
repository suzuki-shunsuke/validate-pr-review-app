# Handling Pull Request Events

Previously, validate-pr-review-app only subscribed to Pull Request Review Events.
Starting with v0.3.2, it can also subscribe to Pull Request Events.
Because only Pull Request Review Events were subscribed to before, pushing empty commits or trivial merge commits to an already-approved PR would not create a validate-pr-review-app check on the pushed commit, requiring an additional approval and degrading the developer experience.
By subscribing to Pull Request Events, validate-pr-review-app creates a check on the pushed commit without requiring an additional approval.
validate-pr-review-app only handles the `synchronize` action of Pull Request Events and ignores all other actions.
If the target commit has reviews, the reviews are validated using the same logic as before.
If there are no reviews and the target commit is neither an empty commit nor a trivial merge commit, no check is created.
[See Allow Empty Commits and Trivial Merge Commits for details about empty commits and trivial merge commits.](allow-empty-commit-and-trivial-merge-commit.md)

If the target commit is an empty commit or a trivial merge commit, validate-pr-review-app walks back through the parent commits on the PR's head branch using the same logic, and creates a check on the target commit.

Here is a more concrete explanation.

1. Suppose a PR containing commit 1 has been approved.

```
commit 1: approved <- HEAD
```

2. Performing an "update branch" and generating commit 2.

```
commit 1: approved
commit 2: update branch (trivial merge commit) <- HEAD
```

validate-pr-review-app checks commit 2.
Since commit 2 has no approvals and is a trivial merge commit, it checks the parent commit 1.
Because commit 1 is approved, validate-pr-review-app marks commit 2 as success.

```
commit 1: approved
commit 2: success <- HEAD
```

3. Pushing an empty commit 3.

```
commit 1: approved
commit 2: trivial merge commit
commit 3: empty commit <- HEAD
```

validate-pr-review-app again walks back through the commits to commit 1, and marks commit 3 as success.

4. Suppose the PR has a conflict, which is resolved and the branch is updated.

```
commit 1: approved
commit 2: trivial merge commit
commit 3: empty commit
commit 4: resolve conflict <- HEAD
```

validate-pr-review-app checks commit 4, but since it is neither a trivial merge commit nor an empty commit, no check is created.
Review is required.

5. Pushing an empty commit again.

```
commit 1: approved
commit 2: trivial merge commit
commit 3: empty commit
commit 4: resolve conflict
commit 5: empty commit <- HEAD
```

validate-pr-review-app walks back through the commits to commit 4, but since commit 4 is not approved and is neither a trivial merge commit nor an empty commit, no check is created.

6. Approving commit 5 and merging a different feature branch into the PR's feature branch.

```
commit 1 ~ 4: omitted
commit 5: approved
commit 6: merge other feature branch <- HEAD
```

Although commit 6 is a merge commit, it does not merge the PR's base branch, so no check is created.
