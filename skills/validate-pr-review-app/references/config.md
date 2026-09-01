# Configuration

Configuration consists of non-secrets and **secrets**.
This document describes the non-secret configuration.
[See Secrets for `webhook_secret` and `github_app_private_key`.](secret.md)

The non-secret configuration is a YAML document.
It's passed by the environment variable `CONFIG` (a YAML string) or `CONFIG_FILE` (a file path).
[See Environment Variables for details.](env.md)

## Contents

- [JSON Schema](#json-schema)
- [Example](#example)
- [Trustness](#trustness)
- [Ignoring Repositories](#ignoring-repositories)
- [Allow Unsigned Commits](#allow-unsigned-commits)
- [Customize footer](customize-footer.md)

## JSON Schema

[json-schema/config.json](https://github.com/suzuki-shunsuke/validate-pr-review-app/blob/main/json-schema/config.json)

Add the following comment to the top of your configuration file for input complementation by [YAML Language Server](https://github.com/redhat-developer/yaml-language-server).
[Please see the comment too.](https://github.com/szksh-lab/.github/issues/67#issuecomment-2564960491)

```yaml
# yaml-language-server: $schema=https://raw.githubusercontent.com/suzuki-shunsuke/validate-pr-review-app/main/json-schema/config.json
```

Or pinning version:

```yaml
# yaml-language-server: $schema=https://raw.githubusercontent.com/suzuki-shunsuke/validate-pr-review-app/v0.0.1/json-schema/config.json
```

### Validate Config Using JSON Schema

You can validate your config using JSON Schema and tools such as [ajv-cli](https://ajv.js.org/packages/ajv-cli.html).

```sh
ajv --spec=draft2020 -s json-schema/config.json -d config.yaml
```

## Example

> [!WARNING]
> Please remove `[bot]` from each app name of `trusted_apps`
> :o: `dependabot`
> :x: `dependabot[bot]`

```yaml
# yaml-language-server: $schema=https://raw.githubusercontent.com/suzuki-shunsuke/validate-pr-review-app/main/json-schema/config.json
# Required
app_id: 0000 # GitHub App ID
installation_id: 00000000 # GitHub App Installation ID

# Optional
aws:
  secret_id: validate-pr-review-app # Secret ID in AWS Secrets Manager
  use_lambda_function_url: true # Optional. true when using Lambda Function URL. Default: false
check_name: validate-review # GitHub Check Name that validate-pr-review-app updates. Default: validate-review
log_level: info # debug, info, warn, error. Default: info
trust:
  trusted_apps:
    # Trusted GitHub App names. The suffix `[bot]` should be removed.
    # Default: renovate and dependabot
    # Glob is unavailable
    - renovate
    - dependabot
  untrusted_machine_users:
    # Untrusted Machine User names.
    # Glob is available
    - "*-bot"
    - "!foo-bot" # If the value starts with "!", the machine user is trusted. This syntax is similar to .gitignore's "!"
    - octocat
repositories:
  # Repository specific config
  # Override the root config
  # Only the first element matching the repository is used
  # If no element matches, the root config is used
  - repositories:
      # Glob pattern matching repository full names
      - suzuki-shunsuke/*
    trust:
      trusted_apps:
        # Trusted GitHub App names. The suffix `[bot]` should be removed.
        # Default: renovate and dependabot
        - renovate
        - dependabot
        - suzuki-shunsuke-app
      untrusted_machine_users:
        - "*-bot"
        - bot-*
  - repositories:
      - suzuki-shunsuke/sandbox
    ignored: true # The app ignores events from this repository and creates no check
    trust: {} # trust is required even when the repository is ignored
```

## Ignoring Repositories

Set `ignored: true` on a `repositories` element to make the app skip the matching repositories.
The app ignores their events entirely and creates no check for them.

```yaml
repositories:
  - repositories:
      - suzuki-shunsuke/sandbox
    ignored: true
    trust: {}
```

## Trustness

- **Trusted Apps & Users**: properly managed, cannot be abused.
- **Untrusted Apps & Users**: potentially exploitable (e.g. private keys exposed).

Approvals from untrusted Machine Users and GitHub Apps are ignored, and commits from them require 2 approvals.
[See Validation Rules.](validation-rules.md)

By default:

- `renovate` and `dependabot` are trusted Apps.
- Machine Users are trusted unless configured otherwise.
  - This is because machine users can't be distinguished from normal users without configuration.

There are three levels of trust configuration, in order of priority:

1. Repository configuration
2. Global configuration
3. Default configuration

Each level defines the following settings:

- `trusted_apps`
- `untrusted_machine_users`

For each setting, only the highest-priority configuration is applied.
Lower-priority configurations are ignored rather than merged.
Therefore, if you want to retain values from lower-priority configurations, you must explicitly include them in the higher-priority configuration.

For example, with the following configuration, apps like `dependabot` and `renovate`—which are trusted by default—will become untrusted:

```yaml
trust:
  trusted_apps:
    - suzuki-shunsuke-app
```

If that is undesirable, you need to explicitly include them:

```yaml
trust:
  trusted_apps:
    - dependabot
    - renovate
    - suzuki-shunsuke-app
```

The `trusted_apps` and `untrusted_machine_users` settings are evaluated independently.
This means you can apply repository-level configuration for `trusted_apps` while still using global configuration for `untrusted_machine_users`.

```yaml
trust:
  untrusted_machine_users:
    - "*-bot"

repositories:
  - repositories:
      - suzuki-shunsuke/pinact
    trust:
      trusted_apps:
        - suzuki-shunsuke-app
```

## Allow Unsigned Commits

> [!WARNING]
> This setting isn't recommended in terms of security.

In the real world, sometimes it's hard to enforce signed commits.
So you may want to allow unsigned commits from certain authors.

- `allow_unsigned_commits`: If this is true, all unsigned commits don't require two approvals. By default, this is false.
- `unsigned_commit_apps`: If this is set, commits from these apps don't require two approvals. By default, this is empty.
  - If `allow_unsigned_commits` is true, please don't set this setting. Otherwise, the app fails.
  - Glob is unavailable
  - `[bot]` should be removed from login
- `unsigned_commit_machine_users`: If this is set, commits from these machine users don't require two approvals. By default, this is empty.
  - If `allow_unsigned_commits` is true, please don't set this setting. Otherwise, the app fails.
  - Glob is unavailable

```yaml
insecure:
  allow_unsigned_commits: true
repositories:
  - repositories:
      - suzuki-shunsuke/*
    insecure: # The repository config overrides the root config.
      unsigned_commit_machine_users:
        - "foo-bot"
      unsigned_commit_apps:
        - "foo-bot"
```
