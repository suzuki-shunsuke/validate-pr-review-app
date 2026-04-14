This pull request requires two approvals because:

{{if .SelfApprovers -}}
The following approvers have self-approved this pull request by pushing commits:
{{range $login, $_ := .SelfApprovers}}
- `{{$login}}`
{{- end}}
{{end}}
{{if .UntrustedCommits -}}
The following commits are untrusted, so two approvals are required.
{{range .UntrustedCommits}}
- {{.SHA}} {{.Login}} {{.Message -}}
{{end -}}
{{end}}
{{if .IgnoredApprovers}}
## :warning: Some approvals are ignored

Approvals from GitHub Apps and Untrusted Machine Users are ignored.

Approvals from the following approvers are ignored:
{{- range .IgnoredApprovers}}
- {{.Login}} {{if .IsApp}}GitHub App{{else}}Untrusted Machine User{{end -}}
{{end}}
{{end}}
{{template "settings" .}}
{{template "footer" . -}}
