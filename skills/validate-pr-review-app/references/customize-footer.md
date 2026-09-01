# Customize footer

You can customize the footer of this app's Checks tab.

The default is: [footer.md](https://github.com/suzuki-shunsuke/validate-pr-review-app/blob/main/pkg/config/templates/footer.md)

For example, you can add the guide for developers:

```yaml
templates:
  footer: |
    ---

    For more details, see the [document](https://example.com).
    If you have any questions, please contact the security team.
```

This template is rendered with [Go's html/template](https://pkg.go.dev/html/template).
