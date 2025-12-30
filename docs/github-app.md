# GitHub App Settings

[Registering a GitHub App](https://docs.github.com/en/apps/creating-github-apps/registering-a-github-app/registering-a-github-app)

- Enable Webhook
- Permissions:
  - Checks: Read and write
  - Contents: Read-only
  - Pull requests: Read-only
- `Where can this GitHub App be installed?` > `Only on this account`
- Install apps into repositories
- [Create a private key](https://docs.github.com/en/apps/creating-github-apps/authenticating-with-a-github-app/managing-private-keys-for-github-apps)
- Subscribe Events
  - Pull request review

After registering the app, you can get the app id from the setting page.
Please add it to config.yaml.

```yaml
app_id: 0123456
```

After installing the app, you can get the installation id from URL.
Please add it to config.yaml.

```yaml
installation_id: 01234567
```
